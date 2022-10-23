create table accounts
(
    id        serial
        constraint accounts_pk
            primary key,
    name      varchar(25)              default ''::character varying not null,
    email     varchar                  default ''::character varying not null,
    secure    varchar                  default ''::character varying not null,
    password  varchar,
    entropy   bytea,
    sample    jsonb                    default '[]'::jsonb           not null,
    rules     jsonb                    default '[]'::jsonb           not null,
    status    boolean                  default false                 not null,
    create_at timestamp with time zone default CURRENT_TIMESTAMP
);

alter table accounts
    owner to envoys;

create unique index accounts_email_uindex
    on accounts (email);

create unique index accounts_id_uindex
    on accounts (id);

INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (1, 'goodman', 'aigarnetwork@gmail.com', '', 'bnHvxWP8NyX1Pgm9mrzQRHkMYwqcDp18jcimJP4oQNI=', E'\\xFDF03535BE97BED878F109DC3412534E', '[]', true, '2022-06-21 15:30:31.221068 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (2, 'Eatrin', 'Eatrin88@outlook.com', '', 'uwxCWTvwlwyHguXps0By0sx1-po15BqbBFVgL31qDeo=', E'\\x2C9AB6BC8557B8B90373EE8E0FC58702', '[]', true, '2022-07-20 09:27:43.630222 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (3, 'Александр', 'ssashakravchuk@gmail.com', '', 'EdsRsFy0oRmmV1o3d1i8C17j77ovDsWEi1gVSRc8NQE=', E'\\xE7239EF774EA4BEAAF494D3664555F29', '[]', true, '2022-07-26 12:54:59.325700 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (4, 'Александр Михайлов', 'alexpro401@gmail.com', '', 'bnHvxWP8NyX1Pgm9mrzQRHkMYwqcDp18jcimJP4oQNI=', E'\\x423876F9AD055E1398AAB8CB54CE6743', '[]', true, '2022-08-02 15:03:54.729619 +00:00', '["currencies", "chains", "pairs", "accounts", "deny-record", "contracts"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (5, 'papaRimskiy', 'dibayok569@bongcs.com', '', '533E4PWiqCjHt7t-xCRq4hV8EeinsKtJ3Ble_V89xP0=', E'\\x01499131704BFE69181849165B211AB3', '[]', true, '2022-09-27 06:45:27.874286 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (6, 'Дядя Жора', 'paym.sssss@gmail.com', '', 'vUPtjVOPvsL2-TIoWDioSnIg1WFWMbYEL9rQVgO8oLE=', E'\\xF2304CC95EDDD28783110EC054C96615', '[]', true, '2022-08-14 15:22:09.304902 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (7, 'Ismail', 'ismailnidjat@gmail.com', '', 'cZFWpKvo5R8ADCs0n93SCmhh9izTqPaBbyUH_f30KLA=', E'\\xBBBD7115B40C084703F3B0C8365475BF', '[]', true, '2022-07-27 06:49:24.253192 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (8, 'Алмаз', 'ashabdanov@gmail.com', '', '1JA2FJRx968_Nh90TZp9V7gdzVEzBuFr0YMG_Fg9xV8=', E'\\x7EF503960DDB0CB43336F2F54699AD3A', '[]', true, '2022-10-10 09:00:29.265412 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (9, 'qwerty', 'jotey78758@ishyp.com', '', '533E4PWiqCjHt7t-xCRq4hV8EeinsKtJ3Ble_V89xP0=', E'\\xEE759D5E184A7586A60BFEACDB872ABD', '[]', true, '2022-09-29 10:26:26.377088 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (10, 'Arslan', 'aturdaliev22@gmail.com', '', 'EsuWU1W3PZPkWm8slRflrjKf6WclOdzBH4FSTS9KxyA=', E'\\x399EA639F1F16D5A7B0B096510BEFD88', '[]', false, '2022-07-27 03:52:21.050145 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (11, 'Nursultan ', 'nursultanmatkaziev@gmail.com', '', 'NfuUcnzOUIt1SDDBedDQwCUv9DD0UjXbzfn7HcG6ZaM=', E'\\xBB91EB48171657B899BF2C07BDD92C6A', '[]', true, '2022-07-27 06:46:04.980950 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (12, 'Nursultan M', 'whitebus321@gmail.com', '', '54hGXRe8nMWidYoR3kduphHNLCIacvQndeCx8AXv4L0=', E'\\x7641B154E8EA1A30337AD08DCE88FDFE', '[]', true, '2022-10-10 09:02:11.810563 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (13, 'Aleks Konot', 'sdfsdfsddsf@gmail.com', '', 'vUPtjVOPvsL2-TIoWDioSnIg1WFWMbYEL9rQVgO8oLE=', E'\\x25C73FF5F746822C361623403E6FCD2E', '["login", "news"]', true, '2022-07-26 12:53:36.105267 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (14, 'dexis', 'nosecor617@ukgent.com', '', 'dA1pS9BF1u_IpZf2RxMbdrVDyp7hQQ28Gcvi5MMJT5Q=', E'\\x08A606A4AD33446F86D48214FFED29EF', '[]', false, '2022-08-11 13:42:16.633402 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (15, 'Dmitry Comarov', 'aleksandekonot25@gmail.com', '', 'vUPtjVOPvsL2-TIoWDioSnIg1WFWMbYEL9rQVgO8oLE=', E'\\xAFC2F7A6F341B6341E2AA2646EB21615', '[]', true, '2022-06-05 12:36:36.493874 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (16, 'Aleksandr Konotopskiy', 'paymex.center@gmail.com', '', 'METU-dMG21lHzumRUuGkF_6WOF6hqXmPz3XxWl7Q4x4=', E'\\x34C9C7411729A47624E50DEA7E0DF7D5', '["news"]', true, '2022-06-05 12:33:37.234748 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "listing", "news", "support", "advertising"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (17, 'qwerty', 'koyid53367@cebaike.com', '', 'blEXHj49mEKDGp4oeSJ1v_FtW1yAYevw03KvhGeyyhI=', E'\\x818606B34C7D69EA6553E652B29D7CD7', '[]', true, '2022-10-03 03:28:47.397879 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (18, 'arsgay', 'fawite5843@ploneix.com', '', '533E4PWiqCjHt7t-xCRq4hV8EeinsKtJ3Ble_V89xP0=', E'\\xDD02A1BC9AA217BCAC86D3FD3E49F45B', '[]', true, '2022-09-27 08:37:22.238933 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (19, 'Almaz', 'a.melisbekov@rpg.kg', '', 's2MWJ2AP8gShYzBaDK-TF4q-4W7kfD0nmXHPke5WpR4=', E'\\x1B33967BE215E2AB179985653630FF58', '[]', true, '2022-07-26 11:52:58.487291 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (20, 'Arstanbek', 'a.abdikaparov@rpg.kg', '', 'aoS0vSzWQIaldzSSbnxLr2eqHffL0cEqO4gPeBajBKQ=', E'\\x77AEEC6ECE6C7F792FC122FD81053885', '[]', true, '2022-07-26 11:43:14.701753 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, status, create_at, rules) VALUES (21, 'Alekdsds', 'admins@envoys.vision', '', 'ehpDFRC5ieU2GNznH92qbigGOHCBGprrLp-TZGgj2D8=', E'\\x2B7463D82A1FB10051FCD50DA8B7B0F2', '["news", "login", "withdrawal", "order_filled"]', true, '2022-08-03 03:55:39.553010 +00:00', '["currencies", "chains", "pairs", "accounts", "contracts", "deny-record"]');

SELECT pg_catalog.setval('public.accounts_id_seq', 21, true);