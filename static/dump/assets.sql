create table assets
(
    id      serial
        constraint assets_pk
            primary key,
    user_id integer,
    symbol  varchar,
    balance numeric(32, 16) default 0.0000000000000000 not null
);

alter table assets
    owner to envoys;

create unique index assets_id_uindex
    on assets (id);
