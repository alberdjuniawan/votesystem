import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const BASE_URL = __ENV.VOTE_API_BASE_URL || 'http://localhost:8080';

const roomCreateRate = new Rate('room_create_rate');
const voteCastRate = new Rate('vote_cast_rate');
const voteDuration = new Trend('vote_duration');

export const options = {
    stages: [
        { duration: '30s', target: 5 },
        { duration: '1m', target: 10 },
        { duration: '2m', target: 10 },
        { duration: '30s', target: 0 },
    ],
    thresholds: {
        http_req_duration: ['p(95)<1000'],
        http_req_failed: ['rate<0.05'],
    },
};

function randomString(len) {
    const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
    let out = '';
    for (let i = 0; i < len; i++) out += chars[Math.floor(Math.random() * chars.length)];
    return out;
}

export default function () {
    const suffix = `${__VU}-${__ITER}-${Date.now()}`;
    const email = `user-${suffix}@test.com`;
    const password = 'password123';
    const name = `User ${__VU}`;
    let token = '';
    let roomId = '';
    let optionIds = [];

    group('Auth', function () {
        let res = http.post(`${BASE_URL}/api/v1/auth/register`, JSON.stringify({
            email: email,
            password: password,
            name: name,
        }), { headers: { 'Content-Type': 'application/json' }, tags: { name: 'Register' } });

        let ok = check(res, {
            'register OK': (r) => r.status === 201 || r.status === 409,
        });
        if (!ok) return;

        res = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
            email: email,
            password: password,
        }), { headers: { 'Content-Type': 'application/json' }, tags: { name: 'Login' } });

        ok = check(res, {
            'login OK': (r) => r.status === 200,
            'has access_token': (r) => r.json() && r.json().access_token !== undefined,
        });
        if (!ok) return;

        token = res.json().access_token;
    });

    if (!token) return;

    const authHeaders = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
    };

    group('Create Room', function () {
        const res = http.post(`${BASE_URL}/api/v1/rooms`, JSON.stringify({
            title: `Test Room ${suffix}`,
            description: 'Created by k6 load test',
            type: Math.random() > 0.5 ? 'single_choice' : 'multiple_choice',
            show_realtime: true,
            max_votes: 1,
        }), { headers: authHeaders, tags: { name: 'CreateRoom' } });

        const ok = check(res, {
            'create room OK': (r) => r.status === 201,
            'has room id': (r) => r.json() && r.json().id !== undefined,
        });

        if (ok) {
            roomId = res.json().id;
            roomCreateRate.add(1);
        }
    });

    if (!roomId) return;

    sleep(0.5);

    group('Add Options', function () {
        const labels = ['Option A', 'Option B', 'Option C', 'Option D'];
        for (const label of labels) {
            const res = http.post(`${BASE_URL}/api/v1/rooms/${roomId}/options`, JSON.stringify({
                label: label,
                description: `Description for ${label}`,
                order_num: labels.indexOf(label),
            }), { headers: authHeaders, tags: { name: 'CreateOption' } });

            const ok = check(res, {
                'create option OK': (r) => r.status === 201,
                'has option id': (r) => r.json() && r.json().id !== undefined,
            });

            if (ok) {
                optionIds.push(res.json().id);
            }
            sleep(0.2);
        }
    });

    group('Activate Room', function () {
        const res = http.patch(`${BASE_URL}/api/v1/rooms/${roomId}/status`, JSON.stringify({
            status: 'active',
        }), { headers: authHeaders, tags: { name: 'ActivateRoom' } });

        check(res, {
            'activate room OK': (r) => r.status === 200,
        });
    });

    sleep(0.3);

    if (optionIds.length > 0) {
        group('Cast Vote', function () {
            const start = Date.now();
            const chosen = optionIds[Math.floor(Math.random() * optionIds.length)];
            const res = http.post(`${BASE_URL}/api/v1/rooms/${roomId}/votes`, JSON.stringify({
                option_id: chosen,
            }), { headers: authHeaders, tags: { name: 'CastVote' } });

            const ok = check(res, {
                'cast vote OK': (r) => r.status === 201,
                'has vote id': (r) => r.json() && r.json().id !== undefined,
            });

            if (ok) {
                voteCastRate.add(1);
                voteDuration.add(Date.now() - start);
            }
        });
    }

    group('Leaderboard', function () {
        const res = http.get(`${BASE_URL}/api/v1/rooms/${roomId}/leaderboard`, {
            tags: { name: 'GetLeaderboard' },
        });

        check(res, {
            'leaderboard OK': (r) => r.status === 200,
            'has scores': (r) => r.json() && Array.isArray(r.json().scores),
        });
    });

    group('Cleanup', function () {
        const res = http.del(`${BASE_URL}/api/v1/rooms/${roomId}`, null, {
            headers: authHeaders,
            tags: { name: 'DeleteRoom' },
        });

        check(res, {
            'delete room OK': (r) => r.status === 204,
        });
    });

    sleep(1);
}
