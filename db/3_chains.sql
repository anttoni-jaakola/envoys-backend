create table if not exists public.spot_chains
(
    id            serial
        constraint spot_chains_pk
            primary key
        constraint spot_chains_id_key
            unique,
    name          varchar,
    rpc           varchar,
    rpc_key       varchar          default ''::character varying not null,
    rpc_user      varchar          default ''::character varying not null,
    rpc_password  varchar          default ''::character varying not null,
    block         integer          default 0                     not null,
    network       integer          default 0                     not null,
    explorer_link varchar          default ''::character varying not null,
    platform      integer          default 0                     not null,
    confirmation  integer          default 3                     not null,
    time_withdraw integer          default 1800                  not null,
    fees_withdraw double precision default 0.5                   not null,
    address       varchar          default ''::character varying not null,
    tag           integer,
    parent_symbol varchar          default ''::character varying not null,
    status        boolean          default false                 not null
);

alter table public.spot_chains
    owner to envoys;

create unique index if not exists spot_chains_id_uindex
    on public.spot_chains (id);

create unique index if not exists spot_chains_name_uindex
    on public.spot_chains (name);

insert into public.spot_chains (id, name, rpc, rpc_key, rpc_user, rpc_password, block, network, explorer_link, platform, confirmation, time_withdraw, fees_withdraw, address, tag, parent_symbol, status)
values  (6, 'Visa Gateway', 'https://github.com', '', '', '', 0, 0, '', 3, 0, 10, 0, '', 0, '', false),
        (7, 'MC Gateway', 'https://github.com/', '', '', '', 0, 0, '', 4, 0, 10, 0, '', 0, '', false),
        (1, 'Tron Chain', 'http://127.0.0.1:8090', '', '', '', 811748, 0, 'https://tronscan.org/#/transaction', 2, 5, 30, 1, '', 4, 'trx', false),
        (2, 'Ethereum Chain', 'http://127.0.0.1:8545', '', '', '', 0, 1000, 'https://etherscan.io/tx', 1, 3, 10, 0.001, '', 2, 'eth', false),
        (3, 'Binance Smart Chain', 'https://bsc-dataseed.binance.org', '', '', '', 0, 56, 'https://bscscan.com/tx', 1, 12, 10, 0.0008, '', 3, 'bnb', false),
        (4, 'Bitcoin Chain', 'https://google.com', '', '', '', 0, 0, 'https://www.blockchain.com/btc/tx', 0, 3, 60, 0.0002, '', 1, 'btc', false);