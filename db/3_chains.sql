create table if not exists public.chains
(
    id            serial
        constraint chains_pk
            primary key
        constraint chains_id_key
            unique,
    name          varchar,
    rpc           varchar,
    block         integer          default 0                     not null,
    network       integer          default 0                     not null,
    explorer_link varchar          default ''::character varying not null,
    platform      integer          default 0                     not null,
    confirmation  integer          default 3                     not null,
    time_withdraw integer          default 1800                  not null,
    fees_withdraw double precision default 0.5                   not null,
    tag           integer,
    parent_symbol varchar          default ''::character varying not null,
    status        boolean          default false                 not null
);

alter table public.chains
    owner to envoys;

alter table public.chains
    add unique (id);

create unique index if not exists chains_id_uindex
    on public.chains (id);

create unique index if not exists chains_name_uindex
    on public.chains (name);

insert into public.chains (id, name, rpc, block, network, explorer_link, platform, confirmation, time_withdraw, fees_withdraw, tag, parent_symbol, status)
values  (7, 'MC Gateway', 'https://github.com/', 0, 0, '', 4, 0, 10, 0, '', 0, '', false),
        (3, 'Binance Smart Chain', 'https://bsc-dataseed.binance.org', 0, 56, 'https://bscscan.com/tx', 1, 12, 10, 0.0008, 3, 'bnb', false),
        (6, 'Visa Gateway', 'https://github.com', 0, 0, '', 3, 0, 10, 0, '', 0, '', false),
        (4, 'Bitcoin Chain', 'https://google.com', 0, 0, 'https://www.blockchain.com/btc/tx', 0, 3, 60, 0.0002, 1, 'btc', false),
        (2, 'Ethereum Chain', 'http://127.0.0.1:8545', 0, 1000, 'https://etherscan.io/tx', 1, 3, 10, 0.001, 2, 'eth', false),
        (1, 'Tron Chain', 'http://127.0.0.1:8090', 1034, 0, 'https://tronscan.org/#/transaction', 2, 5, 30, 1, 4, 'trx', true);