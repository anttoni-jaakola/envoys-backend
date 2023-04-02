create table if not exists public.advertising
(
    id      serial
        primary key,
    title   varchar default ''::character varying     not null,
    text    varchar default ''::character varying     not null,
    link    varchar,
    pattern varchar default 'text'::character varying not null
);

alter table public.advertising
    owner to envoys;

alter table public.advertising
    add unique (id);

create unique index if not exists advertising_id_uindex
    on public.advertising (id);

insert into public.advertising (id, title, text, link, pattern)
values (1, '', '', 'https://filmix.ac/series/triller/135902-k-karnival-rou-2019.html', 'sticker'),
       (2, '', '', 'https://filmix.ac/series/triller/135902-k-karnival-rou-2019.html', 'sticker'),
       (4, '', '', 'https://filmix.ac/series/triller/135902-k-karnival-rou-2019.html', 'sticker'),
       (5, '', '', 'https://filmix.ac/series/triller/135902-k-karnival-rou-2019.html', 'sticker'),
       (6, '', '', 'https://filmix.ac/series/triller/135902-k-karnival-rou-2019.html', 'sticker'),
       (8, '', '', 'https://filmix.ac/series/triller/135902-k-karnival-rou-2019.html', 'sticker');

select pg_catalog.setval('public.advertising_id_seq', 6, true);