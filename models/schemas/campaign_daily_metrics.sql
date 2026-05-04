CREATE TABLE campaign_daily_metrics (
  campaign_id VARCHAR(26),
  date        DATE,
  impressions INT NOT NULL DEFAULT 0,
  clicks INT NOT NULL DEFAULT 0,
  conversions INT NOT NULL DEFAULT 0,
  PRIMARY KEY (campaign_id, date)
);