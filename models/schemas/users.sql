CREATE TABLE users (
  id          VARCHAR(26)  PRIMARY KEY,
  name        VARCHAR(255) NOT NULL,
  email       VARCHAR(255) NOT NULL UNIQUE,
  password    VARCHAR(255) NOT NULL,
  phone       CHAR(10)     NOT NULL,
  role        VARCHAR(20)  NOT NULL DEFAULT 'marketer'
              CHECK (role IN ('admin', 'marketer', 'analyst')),
  bio         TEXT,
  picture     VARCHAR(500),
  created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  deleted_at  TIMESTAMPTZ
);