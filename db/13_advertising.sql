create table if not exists public.advertising
(
    id    serial
        primary key,
    title varchar default ''::character varying not null,
    text  varchar default ''::character varying not null,
    link  varchar,
    type  integer default 0                     not null
);

alter table public.advertising
    owner to envoys;

alter table public.advertising
    add unique (id);

create unique index if not exists advertising_id_uindex
    on public.advertising (id);

insert into public.advertising (id, title, text, link, type)
values  (1, '', '', 'https://filmix.ac/series/triller/135902-k-karnival-rou-2019.html', 1),
        (2, '', '', 'https://filmix.ac/series/triller/135902-k-karnival-rou-2019.html', 1),
        (4, '', '', 'https://filmix.ac/series/triller/135902-k-karnival-rou-2019.html', 1),
        (5, '', '', 'https://filmix.ac/series/triller/135902-k-karnival-rou-2019.html', 1),
        (6, '', '', 'https://filmix.ac/series/triller/135902-k-karnival-rou-2019.html', 1),
        (8, '', '', 'https://filmix.ac/series/triller/135902-k-karnival-rou-2019.html', 1);