SET TimeZone= 'Etc/UTC';
show timezone;
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;
SELECT create_hypertable('spot_trades', 'create_at');

alter table accounts
    add unique (id);
alter table actions
    add unique (id);
alter table spot_wallets
    add unique (id);
alter table spot_transfers
    add unique (id);
alter table spot_transactions
    add unique (id);
alter table spot_trades
    add unique (id);
alter table spot_reserves
    add unique (id);
alter table spot_pairs
    add unique (id);
alter table spot_orders
    add unique (id);
alter table spot_currencies
    add unique (id);
alter table spot_chains
    add unique (id);
alter table spot_assets
    add unique (id);