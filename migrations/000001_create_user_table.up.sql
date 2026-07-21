-- Создание таблицы пользователей
CREATE TABLE "User" (
                        "ID" INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
                        "Name" VARCHAR(255) UNIQUE NOT NULL,
                        "AuthKey" VARCHAR(255) NOT NULL
);

-- Базовый индекс для поиска по названию
CREATE INDEX idx_user_name ON "User"("Name");

-- Индекс для поиска по году
CREATE INDEX idx_user_name_auth ON "User"("Name", "AuthKey");