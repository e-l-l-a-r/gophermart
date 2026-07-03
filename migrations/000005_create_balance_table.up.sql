-- Создание таблицы баланса
CREATE TABLE "Balance" (
                            "ID" SERIAL PRIMARY KEY,
                            "UserId" INTEGER NOT NULL references "User"("ID")
                                ON DELETE CASCADE
                                ON UPDATE CASCADE,
                            "Current" FLOAT NOT NULL,
                            "Withdrawn" INTEGER NOT NULL
);