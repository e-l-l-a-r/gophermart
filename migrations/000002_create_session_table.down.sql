-- Откат создания таблицы сессий
DROP INDEX IF EXISTS idx_session_key;
DROP TABLE IF EXISTS "Session";