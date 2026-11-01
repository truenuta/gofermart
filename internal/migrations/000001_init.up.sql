    CREATE TABLE users (
        id            BIGSERIAL PRIMARY KEY,
        login         TEXT UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
    );
    CREATE TABLE orders (
        number      TEXT PRIMARY KEY,
        user_id     BIGINT NOT NULL REFERENCES users(id),
        status      TEXT NOT NULL DEFAULT 'NEW',
        accrual     NUMERIC(12,2),
        uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now()
  );
  CREATE INDEX orders_user_uploaded_idx ON orders (user_id, uploaded_at DESC);
  CREATE TABLE withdrawals (
        id           BIGSERIAL PRIMARY KEY,
        order_number TEXT NOT NULL,
        user_id      BIGINT NOT NULL REFERENCES users(id),
        sum          NUMERIC(12,2) NOT NULL,
        processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
  );
  CREATE INDEX withdrawals_user_time_idx ON withdrawals (user_id, processed_at DESC);