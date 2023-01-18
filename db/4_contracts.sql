create table if not exists public.spot_contracts
(
    id            serial
        constraint spot_contracts_pk
            primary key,
    symbol        varchar,
    chain_id      integer,
    address       varchar,
    fees_withdraw double precision default 0.5 not null,
    protocol      integer          default 0   not null,
    decimals      integer          default 18  not null
);

alter table public.spot_contracts
    owner to envoys;

create unique index if not exists spot_contracts_address_uindex
    on public.spot_contracts (address);

create unique index if not exists spot_contracts_id_uindex
    on public.spot_contracts (id);

insert into public.spot_contracts (id, symbol, chain_id, address, fees_withdraw, protocol, decimals)
values  (1, 'link', 2, '0x514910771af9ca656af840dff83e8264ecf986ca', 2.24, 1, 18),
        (2, 'bnb', 2, '0xB8c77482e45F1F44dE1745F52C74426C631bDD52', 0.01, 1, 18),
        (3, 'aave', 2, '0x7fc66500c84a76ad7e9c93437bfc5ac33e2ddae9', 0.35, 1, 18),
        (4, 'btc', 2, '0x0316eb71485b0ab14103307bf65a021042c6d380', 0.002, 1, 18),
        (5, 'omg', 2, '0xd26114cd6EE289AccF82350c8d8487fedB8A0C07', 0.008, 1, 18),
        (6, 'eth', 3, '0x2170ed0880ac9a755fd29b2688956bd959f933f8', 0.006, 6, 18),
        (7, 'tst', 2, '0x31b130CDFcEDB08E6CcA9a0d02964C9a0722D32E', 0.0024, 1, 18),
        (9, 'usdt', 2, '0xdac17f958d2ee523a2206206994597c13d831ec7', 0.001, 1, 18),
        (10, 'usdt', 1, 'TZ7XYSbGhxPKquYp8reVrNn22g7YS58JXa', 6, 9, 8),
        (8, 'usdt', 3, '0x55d398326f99059ff775485246999027b3197955', 0.0001, 6, 18);