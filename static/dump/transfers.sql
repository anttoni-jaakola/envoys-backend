create table transfers
(
    id         serial
        constraint transfers_pk
            primary key,
    user_id    integer,
    order_id   integer,
    base_unit  varchar,
    quote_unit varchar,
    price      numeric(20, 8),
    quantity   numeric(32, 16),
    assigning  integer,
    fees       double precision,
    create_at  timestamp with time zone default CURRENT_TIMESTAMP not null
);

alter table transfers
    owner to envoys;

create unique index transfers_id_uindex
    on transfers (id);
