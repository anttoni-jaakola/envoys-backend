create table if not exists public.fees
(
    id            serial
        constraint fees_pk
            primary key
        unique,
    user_id   integer,
    "from"    varchar,
    "to"      varchar,
    value     double precision,
    symbol    varchar,
    platform  integer                  default 0                                   not null,
    coated    boolean                  default false                               not null,
    type      integer                  default 0                                   not null,
    create_at timestamp with time zone default CURRENT_TIMESTAMP
);

alter table public.fees
    owner to envoys;

alter table public.fees
    add unique (id);

create unique index fees_id_uindex
    on fees (id);