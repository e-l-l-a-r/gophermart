-- Создание таблицы списаний
CREATE TABLE "Withdraw" (
                         "ID" SERIAL PRIMARY KEY,
                         "UserId" INTEGER NOT NULL references "User"("ID")
                             ON DELETE CASCADE
                             ON UPDATE CASCADE,
                         "Number" VARCHAR(64) UNIQUE NOT NULL,
                         "Sum" FLOAT NOT NULL,
                         "ProcessedAt" timestamp NOT NULL
);
