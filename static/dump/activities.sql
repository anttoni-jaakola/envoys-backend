create table activities
(
    id        serial
        constraint activities_pk
            primary key,
    os        varchar(20),
    device    varchar(10),
    ip        varchar(255),
    user_id   integer,
    browser   jsonb,
    create_at timestamp with time zone default CURRENT_TIMESTAMP
);

alter table activities
    owner to envoys;