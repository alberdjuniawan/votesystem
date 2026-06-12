ALTER TABLE votes DROP CONSTRAINT uq_votes_room_user_option;
ALTER TABLE votes ADD CONSTRAINT uq_votes_room_user UNIQUE (room_id, user_id);