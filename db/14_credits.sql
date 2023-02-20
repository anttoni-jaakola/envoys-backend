create table public.credits
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

alter table public.credits
    owner to envoys;

alter table public.credits
    add unique (id);

create unique index if not exists credits_id_uindex
    on public.credits (id);

