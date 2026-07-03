-- Создание таблицы сессий
CREATE TABLE "Session" (
                        "ID" SERIAL PRIMARY KEY,
                        "UserId" INTEGER NOT NULL references "User"("ID")
                            ON DELETE CASCADE
                            ON UPDATE CASCADE,
                        "SessionKey" uuid UNIQUE NOT NULL DEFAULT gen_random_uuid(),
                        "ExpiredAt" timestamp NOT NULL DEFAULT now() + interval '10 minutes'
);

-- Базовый индекс для поиска по ключу сессии
CREATE INDEX idx_session_key ON "Session"("SessionKey");
