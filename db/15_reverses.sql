create table public.reverses
(
    id        serial
        primary key,
    user_id   integer,
    "from"    varchar,
    "to"      varchar,
    value     double precision,
    symbol    varchar,
    platform  integer                  default 0 not null,
    create_at timestamp with time zone default CURRENT_TIMESTAMP
);

alter table public.reverses
    owner to envoys;

alter table public.reverses
    add unique (id);

create unique index if not exists reverses_id_uindex
    on public.reverses (id);
