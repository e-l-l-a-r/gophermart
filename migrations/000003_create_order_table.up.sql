-- Создание таблицы заказов
CREATE TABLE "Order" (
                           "ID" SERIAL PRIMARY KEY,
                           "UserId" INTEGER NOT NULL references "User"("ID")
                               ON DELETE CASCADE
                               ON UPDATE CASCADE,
                           "Number" VARCHAR(64) UNIQUE NOT NULL,
                           "Status" INTEGER NOT NULL,
                           "Accrual" FLOAT NOT NULL,
                           "UploadedAt" timestamp NOT NULL
);

CREATE INDEX idx_order_num ON "Order"("Number");