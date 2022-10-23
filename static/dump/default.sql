SET TimeZone= 'Etc/UTC';
show timezone;

SELECT create_hypertable('trades', 'create_at');