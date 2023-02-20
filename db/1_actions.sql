create table if not exists public.actions
(
    id        serial
        constraint actions_pk
            primary key
        constraint actions_id_key
            unique,
    os        varchar(20),
    device    varchar(10),
    ip        varchar(255),
    user_id   integer,
    browser   jsonb,
    create_at timestamp with time zone default CURRENT_TIMESTAMP
);

alter table public.actions
    owner to envoys;

alter table public.actions
    add unique (id);

create unique index if not exists actions_id_uindex
    on public.actions (id);