create table if not exists public.transactions
(
    id           serial,
    symbol       varchar,
    hash         varchar                  default gen_random_uuid()             not null,
    value        numeric(32, 18),
    fees         double precision         default 0                             not null,
    confirmation integer                  default 0                             not null,
    "to"         varchar                  default ''::character varying         not null,
    block        integer                  default 0                             not null,
    chain_id     integer                  default 0                             not null,
    user_id      integer,
    repayment    boolean                  default false                         not null,
    price        double precision         default 0                             not null,
    parent       integer                  default 0                             not null,
    assignment   varchar                  default 'deposit'::character varying  not null,
    "group"      varchar                  default 'crypto'::character varying   not null,
    platform     varchar                  default 'ethereum'::character varying not null,
    protocol     varchar                  default 'erc20'::character varying    not null,
    allocation   varchar                  default 'external'::character varying not null,
    status       varchar                  default 'pending'::character varying  not null,
    error        varchar                  default ''::character varying  not null,
    create_at    timestamp with time zone default CURRENT_TIMESTAMP
);

alter table public.transactions
    owner to envoys;

create unique index if not exists transactions_id_uindex
    on public.transactions (id);

create unique index if not exists transactions_hash_uindex
    on public.transactions (hash);

alter table public.transactions
    add constraint transactions_pk
        primary key (id);

alter table public.transactions
    add unique (id);

alter table public.transactions
    add unique (hash);