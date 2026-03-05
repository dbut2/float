-- name: UpsertFXRate :exec
INSERT INTO float.fx_rates (base_currency, quote_currency, rate, date)
VALUES ($1, $2, $3, $4)
ON CONFLICT (base_currency, quote_currency, date) DO UPDATE SET rate = EXCLUDED.rate;

-- name: GetFXRate :one
SELECT rate FROM float.fx_rates
WHERE base_currency = $1 AND quote_currency = $2 AND date = $3::DATE;
