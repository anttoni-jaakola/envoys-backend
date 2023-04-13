create table if not exists public.kyc
(
    user_id integer
        constraint kyc_user_id_key
            unique,
    secure  boolean default false                        not null,
    secret  varchar default ''::character varying        not null,
    process boolean default false                        not null,
    level   varchar default 'level_0'::character varying not null
);

alter table public.kyc
    owner to envoys;