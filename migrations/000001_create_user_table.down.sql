-- Откат создания таблицы пользователей
DROP INDEX IF EXISTS idx_user_name_auth;
DROP INDEX IF EXISTS idx_user_name;
DROP TABLE IF EXISTS "User";