create table if not exists public.agents
(
    id        serial,
    type      varchar                  default 'agent'::character varying   not null,
    user_id   integer,
    status    varchar                  default 'pending'::character varying not null,
    name      varchar                  default ''::character varying        not null,
    broker_id integer,
    create_at timestamp with time zone default CURRENT_TIMESTAMP
);

alter table public.agents
    owner to envoys;

alter table public.agents
    add constraint agents_pkey
        primary key (id);

alter table public.agents
    add constraint agents_user_id_key
        unique (user_id);