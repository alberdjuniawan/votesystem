import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter, Gauge } from 'k6/metrics';
import { SharedArray } from 'k6/data';

const BASE_URL         = __ENV.VOTE_API_BASE_URL   || 'http://localhost:8080';
const WS_URL           = __ENV.VOTE_WS_BASE_URL    || BASE_URL.replace(/^http/, 'ws');
const CREATOR_VU_COUNT = parseInt(__ENV.K6_CREATOR_VU_COUNT  || '3');
const ROOMS_PER_CREATOR= parseInt(__ENV.K6_ROOMS_PER_CREATOR || '2');

const roomCreateRate   = new Rate('room_create_success');
const voteCastRate     = new Rate('vote_cast_success');
const wsConnectRate    = new Rate('ws_connect_success');
const roomCloseRate    = new Rate('room_close_success');
const roomDeleteRate   = new Rate('room_delete_success');
const voteDuration     = new Trend('vote_duration_ms');
const wsMessageCount   = new Counter('ws_messages_received');
const activeWsConns    = new Gauge('ws_active_connections');

export const options = {
    scenarios: {
        creators: {
            executor: 'constant-vus',
            vus: CREATOR_VU_COUNT,
            duration: '4m',
            tags: { scenario: 'creator' },
            env: { ROLE: 'creator' },
        },
        voters: {
            executor: 'ramping-vus',
            startTime: '15s',
            stages: [
                { duration: '30s', target: 10 },
                { duration: '2m',  target: 25 },
                { duration: '1m',  target: 25 },
                { duration: '30s', target: 0  },
            ],
            tags: { scenario: 'voter' },
            env: { ROLE: 'voter' },
        },
        watchers: {
            executor: 'constant-vus',
            vus: 5,
            startTime: '20s',
            duration: '3m30s',
            tags: { scenario: 'watcher' },
            env: { ROLE: 'watcher' },
        },
    },
    thresholds: {
        http_req_duration:            ['p(95)<1500'],
        http_req_failed:              ['rate<0.05'],
        vote_cast_success:            ['rate>0.85'],
        ws_connect_success:           ['rate>0.90'],
        room_close_success:           ['rate>0.90'],
        room_delete_success:          ['rate>0.90'],
    },
};

function randomString(len) {
    const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
    let out = '';
    for (let i = 0; i < len; i++) out += chars[Math.floor(Math.random() * chars.length)];
    return out;
}

function makeEmail(prefix) {
    return `${prefix}-${__VU}-${__ITER}-${randomString(6)}@test.com`;
}

function jsonHeader(token) {
    const h = { 'Content-Type': 'application/json' };
    if (token) h['Authorization'] = `Bearer ${token}`;
    return h;
}

function registerAndLogin(emailPrefix, password = 'password123') {
    const email = makeEmail(emailPrefix);
    const name  = `User VU${__VU}`;

    http.post(
        `${BASE_URL}/api/v1/auth/register`,
        JSON.stringify({ email, password, name }),
        { headers: jsonHeader(), tags: { name: 'Register' } }
    );
    const res = http.post(
        `${BASE_URL}/api/v1/auth/login`,
        JSON.stringify({ email, password }),
        { headers: jsonHeader(), tags: { name: 'Login' } }
    );

    const ok = check(res, {
        'login 200':        (r) => r.status === 200,
        'has access_token': (r) => r.json()?.data?.access_token !== undefined,
    });

    return ok ? { token: res.json().data.access_token, email } : null;
}

function createFullRoom(token, suffix) {
    const headers = jsonHeader(token);
    let roomId    = null;
    let optionIds = [];

    group('Create Room', () => {
        const type = Math.random() > 0.5 ? 'single_choice' : 'multiple_choice';
        const res  = http.post(
            `${BASE_URL}/api/v1/rooms`,
            JSON.stringify({
                title:        `Load Test Room ${suffix}`,
                description:  'k6 multi-voter test',
                type,
                show_realtime: true,
                max_votes:    type === 'multiple_choice' ? 3 : 1,
            }),
            { headers, tags: { name: 'CreateRoom' } }
        );

        const ok = check(res, {
            'create room 201': (r) => r.status === 201,
            'has room id':     (r) => r.json()?.data?.id !== undefined,
        });
        if (ok) {
            roomId = res.json().data.id;
            roomCreateRate.add(1);
        } else {
            roomCreateRate.add(0);
        }
    });

    if (!roomId) return null;

    group('Add Options', () => {
        const labels = ['Option A', 'Option B', 'Option C', 'Option D'];
        for (const [i, label] of labels.entries()) {
            const res = http.post(
                `${BASE_URL}/api/v1/rooms/${roomId}/options`,
                JSON.stringify({ label, description: `Desc ${label}`, order_num: i }),
                { headers, tags: { name: 'CreateOption' } }
            );

            const ok = check(res, {
                'create option 201': (r) => r.status === 201,
                'has option id':     (r) => r.json()?.data?.id !== undefined,
            });
            if (ok) optionIds.push(res.json().data.id);
            sleep(0.1);
        }
    });

    if (optionIds.length === 0) return null;

    group('Activate Room', () => {
        const res = http.patch(
            `${BASE_URL}/api/v1/rooms/${roomId}/status`,
            JSON.stringify({ status: 'active' }),
            { headers, tags: { name: 'ActivateRoom' } }
        );
        check(res, { 'activate 200': (r) => r.status === 200 });
    });

    return { roomId, optionIds };
}

function castVoteAndCheck(token, roomId, optionIds) {
    const headers = jsonHeader(token);

    group('Cast Vote', () => {
        const start   = Date.now();
        const chosen  = optionIds[Math.floor(Math.random() * optionIds.length)];
        const res     = http.post(
            `${BASE_URL}/api/v1/rooms/${roomId}/votes`,
            JSON.stringify({ option_id: chosen }),
            { headers, tags: { name: 'CastVote' } }
        );

        const ok = check(res, {
            'vote 201':    (r) => r.status === 201,
            'vote created or dup': (r) => r.status === 201 || r.status === 409,
        });
        voteCastRate.add(ok ? 1 : 0);
        if (ok) voteDuration.add(Date.now() - start);
    });

    group('Leaderboard', () => {
        const res = http.get(
            `${BASE_URL}/api/v1/rooms/${roomId}/leaderboard`,
            { tags: { name: 'GetLeaderboard' } }
        );
        check(res, {
            'leaderboard 200': (r) => r.status === 200,
            'has scores':      (r) => Array.isArray(r.json()?.data?.scores),
        });
    });

    group('My Vote', () => {
        const res = http.get(
            `${BASE_URL}/api/v1/rooms/${roomId}/votes/me`,
            { headers, tags: { name: 'GetMyVote' } }
        );
        check(res, {
            'my vote 200': (r) => r.status === 200,
            'has_voted field': (r) => r.json()?.data?.has_voted !== undefined,
        });
    });
}

function watchRoomWS(roomId, listenSeconds = 10) {
    const url = `${WS_URL}/api/v1/ws/rooms/${roomId}`;

    console.log(`[WS] Connecting to ${url}`);

    const res = ws.connect(url, {}, (socket) => {
        activeWsConns.add(1);

        socket.on('open', () => {
            console.log(`[WS] Connected room=${roomId}`);
        });

        socket.on('message', (data) => {
            wsMessageCount.add(1);
            try {
                const msg = JSON.parse(data);
                const eventType = msg.event || msg.type || 'unknown';
                console.log(`[WS] room=${roomId} event=${eventType} payload=${JSON.stringify(msg.data || msg).substring(0, 120)}`);
            } catch (_) {
                console.log(`[WS] room=${roomId} raw_msg=${String(data).substring(0, 120)}`);
            }
        });

        socket.on('error', (e) => {
            console.error(`[WS] Error room=${roomId} err=${e}`);
            activeWsConns.add(-1);
        });

        socket.on('close', (code) => {
            console.log(`[WS] Closed room=${roomId} code=${code}`);
            activeWsConns.add(-1);
        });

        socket.setTimeout(() => {
            console.log(`[WS] Timeout reached, closing room=${roomId}`);
            socket.close(1000);
        }, listenSeconds * 1000);
    });

    const ok = check(res, {
        'ws handshake 101': (r) => r && r.status === 101,
    });

    if (!ok) {
        console.error(`[WS] Handshake FAILED room=${roomId} status=${res ? res.status : 'null'} body=${res ? String(res.body).substring(0, 200) : 'n/a'}`);
    }

    wsConnectRate.add(ok ? 1 : 0);
    return ok;
}

function closeAndDeleteRoom(token, roomId) {
    const headers = jsonHeader(token);

    group('Close Room', () => {
        const res = http.patch(
            `${BASE_URL}/api/v1/rooms/${roomId}/status`,
            JSON.stringify({ status: 'closed' }),
            { headers, tags: { name: 'CloseRoom' } }
        );
        const ok = check(res, {
            'close room 200': (r) => r.status === 200,
        });
        roomCloseRate.add(ok ? 1 : 0);
    });

    group('Delete Room', () => {
        const res = http.del(
            `${BASE_URL}/api/v1/rooms/${roomId}`,
            null,
            { headers, tags: { name: 'DeleteRoom' } }
        );
        const ok = check(res, {
            'delete room 204': (r) => r.status === 204,
        });
        roomDeleteRate.add(ok ? 1 : 0);
    });
}

export function setup() {
    const sharedRooms = [];

    console.log('[setup] Creating shared rooms for voter pool...');

    for (let i = 0; i < CREATOR_VU_COUNT * ROOMS_PER_CREATOR; i++) {
        const password = 'setuppass123';
        const email    = `setup-creator-${i}-${randomString(8)}@test.com`;
        const name     = `Setup Creator ${i}`;

        http.post(
            `${BASE_URL}/api/v1/auth/register`,
            JSON.stringify({ email, password, name }),
            { headers: jsonHeader() }
        );

        const loginRes = http.post(
            `${BASE_URL}/api/v1/auth/login`,
            JSON.stringify({ email, password }),
            { headers: jsonHeader() }
        );

        if (loginRes.status !== 200) continue;

        const token = loginRes.json()?.data?.access_token;
        if (!token) continue;

        const room = createFullRoom(token, `setup-${i}`);
        if (room) {
            sharedRooms.push(room);
            console.log(`[setup] Room created: ${room.roomId} (${room.optionIds.length} options)`);
        }
        sleep(0.3);
    }

    console.log(`[setup] Total shared rooms ready: ${sharedRooms.length}`);
    return { sharedRooms };
}

export default function (data) {
    const { sharedRooms } = data;
    const role = __ENV.ROLE || 'voter';

    const roomPool      = sharedRooms.length > 0 ? sharedRooms : null;
    const poolRoom      = roomPool
        ? roomPool[(__VU + __ITER) % roomPool.length]
        : null;

    if (role === 'creator') {
        const auth = registerAndLogin('creator');
        if (!auth) return;

        const suffix = `${__VU}-${__ITER}`;
        const room   = createFullRoom(auth.token, suffix);
        if (!room) return;

        sleep(0.5);

        const voteCount = Math.floor(Math.random() * 2) + 1;
        for (let v = 0; v < voteCount; v++) {
            castVoteAndCheck(auth.token, room.roomId, room.optionIds);
            sleep(0.3);
        }

        if (Math.random() < 0.4 && room) {
            watchRoomWS(room.roomId, 5);
        }

        closeAndDeleteRoom(auth.token, room.roomId);

        sleep(Math.random() * 2 + 1);
        return;
    }

    if (role === 'watcher') {
        const auth = registerAndLogin('watcher');
        if (!auth) return;

        if (poolRoom) {
            group('Pre-WS Leaderboard', () => {
                const res = http.get(
                    `${BASE_URL}/api/v1/rooms/${poolRoom.roomId}/leaderboard`,
                    { tags: { name: 'GetLeaderboard' } }
                );
                check(res, {
                    'leaderboard 200': (r) => r.status === 200,
                });
            });

            const listenSecs = Math.floor(Math.random() * 10) + 15;
            watchRoomWS(poolRoom.roomId, listenSecs);
        } else {
            const room = createFullRoom(auth.token, `watcher-${__VU}`);
            if (room) watchRoomWS(room.roomId, 10);
        }

        sleep(2);
        return;
    }

    const auth = registerAndLogin('voter');
    if (!auth) return;

    sleep(Math.random() * 1.5);

    if (poolRoom) {
        group('Vote to Shared Room', () => {
            group('Get Room Detail', () => {
                const res = http.get(
                    `${BASE_URL}/api/v1/rooms/${poolRoom.roomId}`,
                    { headers: jsonHeader(auth.token), tags: { name: 'GetRoom' } }
                );
                check(res, { 'get room 200': (r) => r.status === 200 || r.status === 404 });
            });

            group('Get Options', () => {
                const res = http.get(
                    `${BASE_URL}/api/v1/rooms/${poolRoom.roomId}/options`,
                    { tags: { name: 'GetOptions' } }
                );
                check(res, { 'get options 200': (r) => r.status === 200 });
            });

            sleep(0.5);

            castVoteAndCheck(auth.token, poolRoom.roomId, poolRoom.optionIds);
        });

        if (roomPool && roomPool.length > 1 && Math.random() < 0.3) {
            sleep(1);
            const anotherRoom = roomPool[(__VU * 3 + __ITER + 1) % roomPool.length];
            if (anotherRoom && anotherRoom.roomId !== poolRoom.roomId) {
                group('Vote to Another Room', () => {
                    castVoteAndCheck(auth.token, anotherRoom.roomId, anotherRoom.optionIds);
                });
            }
        }

    } else {
        const room = createFullRoom(auth.token, `voter-fallback-${__VU}`);
        if (room) castVoteAndCheck(auth.token, room.roomId, room.optionIds);
    }

    sleep(Math.random() * 2 + 0.5);
}

export function teardown(data) {
    console.log(`[teardown] Test finished. Shared rooms created: ${data.sharedRooms.length}`);
    console.log('[teardown] Creator rooms were automatically closed and deleted during the test.');
    console.log('[teardown] Setup shared rooms are intentionally left in the DB as historical data for the Grafana dashboard.');
}
