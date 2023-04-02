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
    platform      varchar          default 'ethereum'::character varying not null,
    confirmation  integer          default 3                     not null,
    time_withdraw integer          default 1800                  not null,
    fees          double precision default 0.5                   not null,
    tag           varchar          default 'tag_ethereum'::character varying not null,
    parent_symbol varchar          default ''::character varying not null,
    decimals      integer          default 18                    not null,
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

insert into public.chains (id, name, rpc, block, network, explorer_link, platform, confirmation, time_withdraw, fees, tag, parent_symbol, decimals, status)
values  (7, 'MC Gateway', 'https://github.com/', 0, 0, '', 'mastercard', 0, 10, 0, 'tag_none', '', 18, false),
        (3, 'Binance Smart Chain', 'https://bsc-dataseed.binance.org', 0, 56, 'https://bscscan.com/tx', 'ethereum', 12, 10, 0.0008, 'tag_binance', 'bnb', 18, false),
        (6, 'Visa Gateway', 'https://github.com', 0, 0, '', 'visa', 0, 10, 0, 'tag_none', '', 18, false),
        (4, 'Bitcoin Chain', 'https://google.com', 0, 0, 'https://www.blockchain.com/btc/tx', 'bitcoin', 3, 60, 0.0002, 'tag_bitcoin', 'btc', 18, false),
        (2, 'Ethereum Chain', 'http://127.0.0.1:8545', 0, 5000, 'https://etherscan.io/tx', 'ethereum', 3, 10, 0.001, 'tag_ethereum', 'eth', 18, false),
        (1, 'Tron Chain', 'http://127.0.0.1:8090', 0, 0, 'https://tronscan.org/#/transaction', 'tron', 5, 30, 1, 'tag_tron', 'trx', 6, true);

select pg_catalog.setval('public.chains_id_seq', 7, true);