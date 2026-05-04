CREATE TABLE event_logs (
  id           VARCHAR(26)  PRIMARY KEY,
  campaign_id  VARCHAR(26)  NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
  event_type   VARCHAR(20)  NOT NULL
               CHECK (event_type IN ('impression', 'click', 'conversion')),
  source_url   VARCHAR(500),
  ip_address   VARCHAR(45),
  user_agent   VARCHAR(500),
  metadata     JSONB        DEFAULT '{}',
  occurred_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  session_id VARCHAR(64),
  step VARCHAR(50)
);
