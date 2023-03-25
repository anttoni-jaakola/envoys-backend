create table if not exists public.withdraws
(
    id        serial,
    quantity  numeric(20, 18)          default 0.000000000000000000  not null,
    name      varchar                  default ''::character varying not null,
    symbol    varchar,
    user_id   integer,
    broker_id integer,
    status    integer                  default 2                     not null,
    create_at timestamp with time zone default CURRENT_TIMESTAMP,
);

alter table public.withdraws
    owner to envoys;

alter table public.withdraws
    add constraint withdraws_pkey
        primary key (id);