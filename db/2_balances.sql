create table if not exists public.balances
(
    id      serial
        constraint balances_pk
            primary key
        constraint balances_id_key
            unique,
    user_id integer,
    symbol  varchar,
    value   numeric(32, 18) default 0.000000000000000000      not null,
    type    varchar         default 'spot'::character varying not null
);

alter table public.balances
    owner to envoys;

alter table public.balances
    add unique (id);

create unique index if not exists balances_id_uindex
    on public.balances (id);