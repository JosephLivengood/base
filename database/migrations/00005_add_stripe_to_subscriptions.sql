-- +goose Up
-- +goose StatementBegin

ALTER TABLE subscriptions
ADD COLUMN stripe_customer_id TEXT,
ADD COLUMN stripe_subscription_id TEXT,
ADD COLUMN stripe_price_id TEXT,
ADD COLUMN status TEXT NOT NULL DEFAULT 'active';

CREATE INDEX idx_subscriptions_stripe_customer ON subscriptions(stripe_customer_id);
CREATE INDEX idx_subscriptions_stripe_sub ON subscriptions(stripe_subscription_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_subscriptions_stripe_sub;
DROP INDEX IF EXISTS idx_subscriptions_stripe_customer;

ALTER TABLE subscriptions
DROP COLUMN IF EXISTS status,
DROP COLUMN IF EXISTS stripe_price_id,
DROP COLUMN IF EXISTS stripe_subscription_id,
DROP COLUMN IF EXISTS stripe_customer_id;

-- +goose StatementEnd
