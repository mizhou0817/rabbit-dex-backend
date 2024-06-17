-- +goose Up
-- +goose StatementBegin
ALTER TABLE app_market
DROP COLUMN last_trade_price_24high,
DROP COLUMN last_trade_price_24low,
DROP COLUMN average_daily_volume,
DROP COLUMN instant_daily_volume,
DROP COLUMN last_trade_price_24h_change_premium,
DROP COLUMN last_trade_price_24h_change_basis,
DROP COLUMN average_daily_volume_change_premium,
DROP COLUMN average_daily_volume_change_basis;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
