SET TimeZone= 'Etc/UTC';
show timezone;
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;
SELECT create_hypertable('spot_trades', 'create_at');
