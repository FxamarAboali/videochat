UPDATE users SET avatar = replace(avatar, '/storage/public/avatar', '/storage/public/user/avatar');
UPDATE users SET avatar_big = replace(avatar_big, '/storage/public/avatar', '/storage/public/user/avatar');