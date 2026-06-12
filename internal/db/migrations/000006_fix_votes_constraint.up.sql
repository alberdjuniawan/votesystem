ALTER TABLE votes DROP CONSTRAINT uq_votes_room_user;
ALTER TABLE votes ADD CONSTRAINT uq_votes_room_user_option UNIQUE (room_id, user_id, option_id);