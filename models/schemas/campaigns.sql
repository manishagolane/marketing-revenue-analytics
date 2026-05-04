CREATE TABLE campaigns (
  id          VARCHAR(26)   PRIMARY KEY,
  name        VARCHAR(255)  NOT NULL,
  description TEXT,
  created_by  VARCHAR(26)   NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  status      VARCHAR(20)   NOT NULL DEFAULT 'draft'
              CHECK (status IN ('draft', 'active', 'paused', 'completed', 'archived')),
  channel     VARCHAR(50),
  budget      NUMERIC(14,2) NOT NULL DEFAULT 0,
  spend       NUMERIC(14,2) NOT NULL DEFAULT 0,
  revenue     NUMERIC(14,2) NOT NULL DEFAULT 0,
  is_public   BOOLEAN       NOT NULL DEFAULT FALSE,
  starts_at   TIMESTAMPTZ,
  ends_at     TIMESTAMPTZ,
  created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  deleted_at  TIMESTAMPTZ
);