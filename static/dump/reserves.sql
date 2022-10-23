create table reserves
(
    id       serial
        constraint reserves_pk
            primary key,
    user_id  integer,
    address  varchar,
    symbol   varchar,
    platform integer,
    protocol integer,
    value    numeric(32, 16) default 0.0000000000000000 not null,
    lock     boolean         default false              not null
);

alter table reserves
    owner to envoys;

create unique index reserves_id_uindex
    on reserves (id);