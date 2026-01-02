DROP INDEX IF EXISTS idx_outbox_created_at;
DROP INDEX IF EXISTS idx_outbox_status;
DROP TABLE IF EXISTS outbox;

DROP INDEX IF EXISTS idx_withdrawals_user_id;
DROP TABLE IF EXISTS withdrawals;

DROP TABLE IF EXISTS balances;

DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_number;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP TABLE IF EXISTS orders;

DROP INDEX IF EXISTS idx_users_login;
DROP TABLE IF EXISTS users;


