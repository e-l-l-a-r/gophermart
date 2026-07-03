-- Откат создания таблицы заказов
DROP INDEX IF EXISTS idx_order_num;
DROP TABLE IF EXISTS "Order";