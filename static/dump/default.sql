SET TimeZone= 'Etc/UTC';
show timezone;

SELECT create_hypertable('spot_trades', 'create_at');