create table public.fees
(
    id        serial
        primary key,
    user_id   integer,
    address   varchar,
    value     double precision,
    symbol    varchar,
    platform  integer                  default 0     not null,
    coated    boolean                  default false not null,
    create_at timestamp with time zone default CURRENT_TIMESTAMP
);

alter table public.fees
    owner to envoys;

alter table public.fees
    add unique (id);

create unique index if not exists fees_id_uindex
    on public.fees (id);

