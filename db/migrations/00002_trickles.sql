-- +goose Up

CREATE TABLE IF NOT EXISTS float.bucket_trickles (
    trickle_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_bucket_id UUID NOT NULL REFERENCES float.buckets(bucket_id) ON DELETE CASCADE,
    to_bucket_id   UUID NOT NULL REFERENCES float.buckets(bucket_id) ON DELETE CASCADE,
    amount_cents   BIGINT NOT NULL CHECK (amount_cents > 0),
    description    TEXT NOT NULL DEFAULT '',
    period         TEXT NOT NULL CHECK (period IN ('daily', 'weekly', 'fortnightly', 'monthly')),
    start_date     DATE NOT NULL DEFAULT CURRENT_DATE,
    end_date       DATE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (from_bucket_id IS DISTINCT FROM to_bucket_id)
);

-- +goose Down
DROP TABLE IF EXISTS float.bucket_trickles;
