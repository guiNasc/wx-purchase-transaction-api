SELECT 'CREATE DATABASE wx_purchase_transactions'
WHERE NOT EXISTS (
  SELECT FROM pg_database WHERE datname = 'wx_purchase_transactions'
)\gexec

\connect wx_purchase_transactions   

CREATE TABLE IF NOT EXISTS purchase_transactions (
  id BIGSERIAL PRIMARY KEY,
  description VARCHAR(50) NOT NULL,
  amount NUMERIC(18, 6) NOT NULL CHECK (amount > 0),
  reference_date DATE NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_purchase_transactions_reference_date
  ON purchase_transactions (reference_date);
