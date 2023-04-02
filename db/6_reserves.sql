create table if not exists public.reserves
(
    id       serial
        constraint reserves_pk
            primary key
        unique,
    user_id  integer,
    address  varchar,
    symbol   varchar,
    platform varchar         default 'ethereum'::character varying not null,
    protocol varchar         default 'erc20'::character varying    not null,
    value    numeric(32, 18) default 0.000000000000000000          not null,
    reverse  numeric(32, 18) default 0.000000000000000000          not null,
    lock     boolean         default false                         not null
);

alter table public.reserves
    owner to envoys;

alter table public.reserves
    add unique (id);

create unique index if not exists reserves_id_uindex
    on public.reserves (id);;