
CREATE TABLE public.accounts (
    id integer NOT NULL,
    name character varying(25) DEFAULT ''::character varying NOT NULL,
    email character varying DEFAULT ''::character varying NOT NULL,
    secure character varying DEFAULT ''::character varying NOT NULL,
    password character varying,
    entropy bytea,
    sample jsonb DEFAULT '[]'::jsonb NOT NULL,
    rules jsonb DEFAULT '{}'::jsonb NOT NULL,
    status boolean DEFAULT false NOT NULL,
    create_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE public.accounts OWNER TO envoys;

CREATE SEQUENCE public.accounts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE public.accounts_id_seq OWNER TO envoys;
ALTER SEQUENCE public.accounts_id_seq OWNED BY public.accounts.id;

CREATE TABLE public.actions (
    id integer NOT NULL,
    os character varying(20),
    device character varying(10),
    ip character varying(255),
    user_id integer,
    browser jsonb,
    create_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE public.actions OWNER TO envoys;

CREATE SEQUENCE public.actions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.actions_id_seq OWNER TO envoys;
ALTER SEQUENCE public.actions_id_seq OWNED BY public.actions.id;

CREATE TABLE public.spot_assets (
    id integer NOT NULL,
    user_id integer,
    symbol character varying,
    balance numeric(32,16) DEFAULT 0.0000000000000000 NOT NULL
);

ALTER TABLE public.spot_assets OWNER TO envoys;

CREATE SEQUENCE public.spot_assets_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE public.spot_assets_id_seq OWNER TO envoys;
ALTER SEQUENCE public.spot_assets_id_seq OWNED BY public.spot_assets.id;

CREATE TABLE public.spot_chains (
    id integer NOT NULL,
    name character varying,
    rpc character varying,
    rpc_key character varying DEFAULT ''::character varying NOT NULL,
    rpc_user character varying DEFAULT ''::character varying NOT NULL,
    rpc_password character varying DEFAULT ''::character varying NOT NULL,
    block integer DEFAULT 0 NOT NULL,
    network integer DEFAULT 0 NOT NULL,
    explorer_link character varying DEFAULT ''::character varying NOT NULL,
    platform integer DEFAULT 0 NOT NULL,
    confirmation integer DEFAULT 3 NOT NULL,
    time_withdraw integer DEFAULT 1800 NOT NULL,
    fees_withdraw double precision DEFAULT 0.5 NOT NULL,
    address character varying DEFAULT ''::character varying NOT NULL,
    tag integer,
    parent_symbol character varying DEFAULT ''::character varying NOT NULL,
    status boolean DEFAULT false NOT NULL
);

ALTER TABLE public.spot_chains OWNER TO envoys;

CREATE SEQUENCE public.spot_chains_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE public.spot_chains_id_seq OWNER TO envoys;
ALTER SEQUENCE public.spot_chains_id_seq OWNED BY public.spot_chains.id;

CREATE TABLE public.spot_contracts (
    id integer NOT NULL,
    symbol character varying,
    chain_id integer,
    address character varying,
    fees_withdraw double precision DEFAULT 0.5 NOT NULL,
    protocol integer DEFAULT 0 NOT NULL,
    decimals integer DEFAULT 18 NOT NULL
);

ALTER TABLE public.spot_contracts OWNER TO envoys;

CREATE SEQUENCE public.spot_contracts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.spot_contracts_id_seq OWNER TO envoys;
ALTER SEQUENCE public.spot_contracts_id_seq OWNED BY public.spot_contracts.id;

CREATE TABLE public.spot_currencies (
    id integer NOT NULL,
    name character varying(40),
    symbol character varying(6),
    min_withdraw numeric(8,4) DEFAULT 0.0001 NOT NULL,
    max_withdraw numeric(20,8) DEFAULT 100 NOT NULL,
    min_deposit numeric(8,4) DEFAULT 0.0001 NOT NULL,
    min_trade numeric(8,4) DEFAULT 0.0001 NOT NULL,
    max_trade numeric(20,8) DEFAULT 1000000 NOT NULL,
    fees_trade numeric(4,4) DEFAULT 0.1 NOT NULL,
    fees_discount numeric(4,4) DEFAULT 0 NOT NULL,
    fees_charges numeric(32,16) DEFAULT 0.0000000000000000 NOT NULL,
    fees_costs numeric(32,16) DEFAULT 0.0000000000000000 NOT NULL,
    marker boolean DEFAULT false NOT NULL,
    chains jsonb DEFAULT '[]'::jsonb NOT NULL,
    status boolean DEFAULT false NOT NULL,
    fin_type integer DEFAULT 0 NOT NULL,
    create_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE public.spot_currencies OWNER TO envoys;

COMMENT ON COLUMN public.spot_currencies.fin_type IS '0 - crypto
1 - fiat';

CREATE SEQUENCE public.spot_currencies_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE public.spot_currencies_id_seq OWNER TO envoys;
ALTER SEQUENCE public.spot_currencies_id_seq OWNED BY public.spot_currencies.id;

CREATE TABLE public.market_orders (
    symbol character varying,
    uid integer,
    price numeric(32,16) DEFAULT 0.0000000000000000 NOT NULL,
    size numeric(32,16) DEFAULT 0.0000000000000000,
    side character varying,
    type character varying,
    cid integer,
    create_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    id character varying,
    volume numeric(32,16) DEFAULT 0.0000000000000000
);

ALTER TABLE public.market_orders OWNER TO envoys;

CREATE TABLE public.spot_orders (
    id integer NOT NULL,
    assigning integer DEFAULT 0 NOT NULL,
    price numeric(20,8) DEFAULT 0.00000000 NOT NULL,
    value numeric(32,16) DEFAULT 0.0000000000000000 NOT NULL,
    quantity numeric(32,16) DEFAULT 0.0000000000000000 NOT NULL,
    base_unit character varying,
    quote_unit character varying,
    user_id integer,
    type integer DEFAULT 0 NOT NULL,
    status integer DEFAULT 2 NOT NULL,
    create_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE public.spot_orders OWNER TO envoys;

COMMENT ON COLUMN public.spot_orders.assigning IS 'Buy = 0 or Sell = 1 type.';
COMMENT ON COLUMN public.spot_orders.status IS 'CANCEL = 0, FILLED = 1, PENDING = 2';

CREATE SEQUENCE public.spot_orders_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE public.spot_orders_id_seq OWNER TO envoys;
ALTER SEQUENCE public.spot_orders_id_seq OWNED BY public.spot_orders.id;

CREATE TABLE public.spot_pairs (
    id integer NOT NULL,
    base_unit character varying,
    quote_unit character varying,
    price numeric(20,8) DEFAULT 0.00000000 NOT NULL,
    base_decimal integer DEFAULT 2 NOT NULL,
    quote_decimal integer DEFAULT 8 NOT NULL,
    status boolean DEFAULT false NOT NULL
);

ALTER TABLE public.spot_pairs OWNER TO envoys;
CREATE SEQUENCE public.spot_pairs_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.spot_pairs_id_seq OWNER TO envoys;
ALTER SEQUENCE public.spot_pairs_id_seq OWNED BY public.spot_pairs.id;
CREATE TABLE public.spot_reserves (
    id integer NOT NULL,
    user_id integer,
    address character varying,
    symbol character varying,
    platform integer,
    protocol integer,
    value numeric(32,16) DEFAULT 0.0000000000000000 NOT NULL,
    lock boolean DEFAULT false NOT NULL
);

ALTER TABLE public.spot_reserves OWNER TO envoys;

CREATE SEQUENCE public.spot_reserves_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.spot_reserves_id_seq OWNER TO envoys;
ALTER SEQUENCE public.spot_reserves_id_seq OWNED BY public.spot_reserves.id;

CREATE TABLE public.spot_trades (
    id integer NOT NULL,
    base_unit character varying,
    quote_unit character varying,
    price numeric(20,8) DEFAULT 0.00000000 NOT NULL,
    quantity numeric(32,16) DEFAULT 0.0000000000000000 NOT NULL,
    assigning integer DEFAULT 0 NOT NULL,
    market boolean DEFAULT false NOT NULL,
    create_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE public.spot_trades OWNER TO envoys;
CREATE TABLE public.spot_transactions (
    id integer NOT NULL,
    symbol character varying,
    hash character varying DEFAULT ''::character varying NOT NULL,
    value numeric(32,16),
    fees double precision DEFAULT 0 NOT NULL,
    confirmation integer DEFAULT 0 NOT NULL,
    "to" character varying DEFAULT ''::character varying NOT NULL,
    block integer DEFAULT 0 NOT NULL,
    chain_id integer DEFAULT 0 NOT NULL,
    user_id integer,
    tx_type integer DEFAULT 0 NOT NULL,
    fin_type integer DEFAULT 0 NOT NULL,
    platform integer,
    protocol integer,
    claim boolean DEFAULT false NOT NULL,
    price double precision DEFAULT 0 NOT NULL,
    status integer DEFAULT 2 NOT NULL,
    create_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE public.spot_transactions OWNER TO envoys;
CREATE TABLE public.spot_transfers (
    id integer NOT NULL,
    user_id integer,
    order_id integer,
    base_unit character varying,
    quote_unit character varying,
    price numeric(20,8),
    quantity numeric(32,16),
    assigning integer,
    fees double precision,
    create_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE public.spot_transfers OWNER TO envoys;

CREATE TABLE public.spot_wallets (
    id integer NOT NULL,
    address character varying,
    user_id integer,
    platform integer DEFAULT 0 NOT NULL,
    protocol integer DEFAULT 0 NOT NULL,
    symbol character varying
);


ALTER TABLE public.spot_wallets OWNER TO envoys;

CREATE SEQUENCE public.spot_trades_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.spot_trades_id_seq OWNER TO envoys;
ALTER SEQUENCE public.spot_trades_id_seq OWNED BY public.spot_trades.id;

CREATE SEQUENCE public.spot_transactions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.spot_transactions_id_seq OWNER TO envoys;
ALTER SEQUENCE public.spot_transactions_id_seq OWNED BY public.spot_transactions.id;

CREATE SEQUENCE public.spot_transfers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.spot_transfers_id_seq OWNER TO envoys;
ALTER SEQUENCE public.spot_transfers_id_seq OWNED BY public.spot_transfers.id;

CREATE SEQUENCE public.spot_wallets_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.spot_wallets_id_seq OWNER TO envoys;
ALTER SEQUENCE public.spot_wallets_id_seq OWNED BY public.spot_wallets.id;
ALTER TABLE ONLY public.accounts ALTER COLUMN id SET DEFAULT nextval('public.accounts_id_seq'::regclass);
ALTER TABLE ONLY public.actions ALTER COLUMN id SET DEFAULT nextval('public.actions_id_seq'::regclass);
ALTER TABLE ONLY public.spot_assets ALTER COLUMN id SET DEFAULT nextval('public.spot_assets_id_seq'::regclass);
ALTER TABLE ONLY public.spot_chains ALTER COLUMN id SET DEFAULT nextval('public.spot_chains_id_seq'::regclass);
ALTER TABLE ONLY public.spot_contracts ALTER COLUMN id SET DEFAULT nextval('public.spot_contracts_id_seq'::regclass);
ALTER TABLE ONLY public.spot_currencies ALTER COLUMN id SET DEFAULT nextval('public.spot_currencies_id_seq'::regclass);
ALTER TABLE ONLY public.spot_orders ALTER COLUMN id SET DEFAULT nextval('public.spot_orders_id_seq'::regclass);
ALTER TABLE ONLY public.spot_pairs ALTER COLUMN id SET DEFAULT nextval('public.spot_pairs_id_seq'::regclass);
ALTER TABLE ONLY public.spot_reserves ALTER COLUMN id SET DEFAULT nextval('public.spot_reserves_id_seq'::regclass);
ALTER TABLE ONLY public.spot_trades ALTER COLUMN id SET DEFAULT nextval('public.spot_trades_id_seq'::regclass);
ALTER TABLE ONLY public.spot_transactions ALTER COLUMN id SET DEFAULT nextval('public.spot_transactions_id_seq'::regclass);
ALTER TABLE ONLY public.spot_transfers ALTER COLUMN id SET DEFAULT nextval('public.spot_transfers_id_seq'::regclass);
ALTER TABLE ONLY public.spot_wallets ALTER COLUMN id SET DEFAULT nextval('public.spot_wallets_id_seq'::regclass);

INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (1, 'goodman', 'aigarnetwork@gmail.com', '', 'bnHvxWP8NyX1Pgm9mrzQRHkMYwqcDp18jcimJP4oQNI=', '\xfdf03535be97bed878f109dc3412534e', '[]', '{}', true, '2022-06-21 18:30:31.221068+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (2, 'Eatrin', 'Eatrin88@outlook.com', '', 'uwxCWTvwlwyHguXps0By0sx1-po15BqbBFVgL31qDeo=', '\x2c9ab6bc8557b8b90373ee8e0fc58702', '[]', '{}', true, '2022-07-20 12:27:43.630222+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (3, 'Александр', 'ssashakravchuk@gmail.com', '', 'EdsRsFy0oRmmV1o3d1i8C17j77ovDsWEi1gVSRc8NQE=', '\xe7239ef774ea4beaaf494d3664555f29', '[]', '{}', true, '2022-07-26 15:54:59.3257+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (4, 'Александр Михайлов', 'alexpro401@gmail.com', '', 'bnHvxWP8NyX1Pgm9mrzQRHkMYwqcDp18jcimJP4oQNI=', '\x423876f9ad055e1398aab8cb54ce6743', '[]', '{}', true, '2022-08-02 18:03:54.729619+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (5, 'papaRimskiy', 'dibayok569@bongcs.com', '', '533E4PWiqCjHt7t-xCRq4hV8EeinsKtJ3Ble_V89xP0=', '\x01499131704bfe69181849165b211ab3', '[]', '{}', true, '2022-09-27 09:45:27.874286+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (6, 'Дядя Жора', 'paym.sssss@gmail.com', '', 'vUPtjVOPvsL2-TIoWDioSnIg1WFWMbYEL9rQVgO8oLE=', '\xf2304cc95eddd28783110ec054c96615', '[]', '{}', true, '2022-08-14 18:22:09.304902+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (7, 'Ismail', 'ismailnidjat@gmail.com', '', 'cZFWpKvo5R8ADCs0n93SCmhh9izTqPaBbyUH_f30KLA=', '\xbbbd7115b40c084703f3b0c8365475bf', '[]', '{}', true, '2022-07-27 09:49:24.253192+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (8, 'Алмаз', 'ashabdanov@gmail.com', '', '1JA2FJRx968_Nh90TZp9V7gdzVEzBuFr0YMG_Fg9xV8=', '\x7ef503960ddb0cb43336f2f54699ad3a', '[]', '{}', true, '2022-10-10 12:00:29.265412+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (9, 'qwerty', 'jotey78758@ishyp.com', '', '533E4PWiqCjHt7t-xCRq4hV8EeinsKtJ3Ble_V89xP0=', '\xee759d5e184a7586a60bfeacdb872abd', '[]', '{}', true, '2022-09-29 13:26:26.377088+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (10, 'Arslan', 'aturdaliev22@gmail.com', '', 'EsuWU1W3PZPkWm8slRflrjKf6WclOdzBH4FSTS9KxyA=', '\x399ea639f1f16d5a7b0b096510befd88', '[]', '{}', false, '2022-07-27 06:52:21.050145+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (11, 'Nursultan ', 'nursultanmatkaziev@gmail.com', '', 'NfuUcnzOUIt1SDDBedDQwCUv9DD0UjXbzfn7HcG6ZaM=', '\xbb91eb48171657b899bf2c07bdd92c6a', '[]', '{}', true, '2022-07-27 09:46:04.98095+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (12, 'Nursultan M', 'whitebus321@gmail.com', '', '54hGXRe8nMWidYoR3kduphHNLCIacvQndeCx8AXv4L0=', '\x7641b154e8ea1a30337ad08dce88fdfe', '[]', '{}', true, '2022-10-10 12:02:11.810563+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (13, 'Aleks Konot', 'sdfsdfsddsf@gmail.com', '', 'vUPtjVOPvsL2-TIoWDioSnIg1WFWMbYEL9rQVgO8oLE=', '\x25c73ff5f746822c361623403e6fcd2e', '["login", "news"]', '{}', true, '2022-07-26 15:53:36.105267+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (14, 'dexis', 'nosecor617@ukgent.com', '', 'dA1pS9BF1u_IpZf2RxMbdrVDyp7hQQ28Gcvi5MMJT5Q=', '\x08a606a4ad33446f86d48214ffed29ef', '[]', '{}', false, '2022-08-11 16:42:16.633402+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (15, 'Dmitry Comarov', 'aleksandekonot25@gmail.com', '', 'vUPtjVOPvsL2-TIoWDioSnIg1WFWMbYEL9rQVgO8oLE=', '\xafc2f7a6f341b6341e2aa2646eb21615', '[]', '{}', true, '2022-06-05 15:36:36.493874+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (17, 'qwerty', 'koyid53367@cebaike.com', '', 'blEXHj49mEKDGp4oeSJ1v_FtW1yAYevw03KvhGeyyhI=', '\x818606b34c7d69ea6553e652b29d7cd7', '[]', '{}', true, '2022-10-03 06:28:47.397879+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (18, 'arsgay', 'fawite5843@ploneix.com', '', '533E4PWiqCjHt7t-xCRq4hV8EeinsKtJ3Ble_V89xP0=', '\xdd02a1bc9aa217bcac86d3fd3e49f45b', '[]', '{}', true, '2022-09-27 11:37:22.238933+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (19, 'Almaz', 'a.melisbekov@rpg.kg', '', 's2MWJ2AP8gShYzBaDK-TF4q-4W7kfD0nmXHPke5WpR4=', '\x1b33967be215e2ab179985653630ff58', '[]', '{}', true, '2022-07-26 14:52:58.487291+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (20, 'Arstanbek', 'a.abdikaparov@rpg.kg', '', 'aoS0vSzWQIaldzSSbnxLr2eqHffL0cEqO4gPeBajBKQ=', '\x77aeec6ece6c7f792fc122fd81053885', '[]', '{}', true, '2022-07-26 14:43:14.701753+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (21, 'Alekdsds', 'admins@envoys.vision', '', 'ehpDFRC5ieU2GNznH92qbigGOHCBGprrLp-TZGgj2D8=', '\x2b7463d82a1fb10051fcd50da8b7b0f2', '["news", "login", "withdrawal", "order_filled"]', '{}', true, '2022-08-03 06:55:39.55301+03');
INSERT INTO public.accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at) VALUES (16, 'Aleksandr Konotopskiy', 'paymex.center@gmail.com', '', 'METU-dMG21lHzumRUuGkF_6WOF6hqXmPz3XxWl7Q4x4=', '\x34c9c7411729a47624e50dea7e0df7d5', '["news"]', '{"spot": ["currencies", "chains", "pairs", "contracts", "listing"], "default": ["accounts", "news", "support", "advertising"]}', true, '2022-06-05 15:33:37.234748+03');

INSERT INTO public.actions (id, os, device, ip, user_id, browser, create_at)VALUES (1, 'linux', 'desktop', '127.0.0.1', 16, '["chrome", "107.0.0.0"]', '2022-11-25 02:52:46.530983+02');
INSERT INTO public.actions (id, os, device, ip, user_id, browser, create_at)VALUES (2, 'linux', 'desktop', '127.0.0.1', 16, '["chrome", "108.0.0.0"]', '2022-12-02 11:02:03.881996+02');
INSERT INTO public.actions (id, os, device, ip, user_id, browser, create_at)VALUES (3, 'linux', 'desktop', '127.0.0.1', 16, '["chrome", "108.0.0.0"]', '2022-12-02 15:04:05.52597+02');
INSERT INTO public.actions (id, os, device, ip, user_id, browser, create_at)VALUES (4, 'linux', 'desktop', '127.0.0.1', 16, '["chrome", "108.0.0.0"]', '2022-12-08 15:24:36.262159+02');
INSERT INTO public.actions (id, os, device, ip, user_id, browser, create_at)VALUES (5, 'linux', 'desktop', '127.0.0.1', 16, '["chrome", "108.0.0.0"]', '2022-12-09 11:37:23.78959+02');
INSERT INTO public.actions (id, os, device, ip, user_id, browser, create_at)VALUES (6, 'linux', 'desktop', '127.0.0.1', 16, '["chrome", "108.0.0.0"]', '2022-12-10 12:05:35.419422+02');
INSERT INTO public.actions (id, os, device, ip, user_id, browser, create_at)VALUES (7, 'linux', 'desktop', '127.0.0.1', 16, '["chrome", "108.0.0.0"]', '2022-12-10 23:33:05.316878+02');

INSERT INTO public.spot_assets (id, user_id, symbol, balance) VALUES (2, 16, 'usdt', 0.0000000000000000);
INSERT INTO public.spot_assets (id, user_id, symbol, balance) VALUES (1, 16, 'trx', 9001.0000000000000000);
INSERT INTO public.spot_assets (id, user_id, symbol, balance) VALUES (4, 16, 'usd', 0.0000000000000000);
INSERT INTO public.spot_assets (id, user_id, symbol, balance) VALUES (3, 16, 'eth', 1.1250000200000000);

INSERT INTO public.spot_chains (id, name, rpc, rpc_key, rpc_user, rpc_password, block, network, explorer_link, platform, confirmation, time_withdraw, fees_withdraw, address, tag, parent_symbol, status) VALUES (7, 'MC Gateway', 'https://github.com/', '', '', '', 0, 0, '', 4, 0, 10, 0, '', 0, '', false);
INSERT INTO public.spot_chains (id, name, rpc, rpc_key, rpc_user, rpc_password, block, network, explorer_link, platform, confirmation, time_withdraw, fees_withdraw, address, tag, parent_symbol, status) VALUES (3, 'Binance Smart Chain', 'https://bsc-dataseed.binance.org', '', '', '', 0, 56, 'https://bscscan.com/tx', 1, 12, 10, 0.0008, '', 3, 'bnb', false);
INSERT INTO public.spot_chains (id, name, rpc, rpc_key, rpc_user, rpc_password, block, network, explorer_link, platform, confirmation, time_withdraw, fees_withdraw, address, tag, parent_symbol, status) VALUES (2, 'Ethereum Chain', 'http://127.0.0.1:8545', '', '', '', 471, 5777, 'https://etherscan.io/tx', 1, 3, 10, 0.001, '', 2, 'eth', false);
INSERT INTO public.spot_chains (id, name, rpc, rpc_key, rpc_user, rpc_password, block, network, explorer_link, platform, confirmation, time_withdraw, fees_withdraw, address, tag, parent_symbol, status) VALUES (1, 'Tron Chain', 'http://127.0.0.1:8090', '', '', '', 3771, 0, 'https://tronscan.org/#/transaction', 2, 5, 30, 1, '', 4, 'trx', false);
INSERT INTO public.spot_chains (id, name, rpc, rpc_key, rpc_user, rpc_password, block, network, explorer_link, platform, confirmation, time_withdraw, fees_withdraw, address, tag, parent_symbol, status) VALUES (4, 'Bitcoin Chain', 'https://google.com', '', '', '', 0, 0, 'https://www.blockchain.com/btc/tx', 0, 3, 60, 0.0002, '', 1, 'btc', false);
INSERT INTO public.spot_chains (id, name, rpc, rpc_key, rpc_user, rpc_password, block, network, explorer_link, platform, confirmation, time_withdraw, fees_withdraw, address, tag, parent_symbol, status) VALUES (6, 'Visa Gateway', 'https://github.com', '', '', '', 0, 0, '', 3, 0, 10, 0, '', 0, '', false);

INSERT INTO public.spot_contracts (id, symbol, chain_id, address, fees_withdraw, protocol, decimals) VALUES (1, 'link', 2, '0x514910771af9ca656af840dff83e8264ecf986ca', 2.24, 1, 18);
INSERT INTO public.spot_contracts (id, symbol, chain_id, address, fees_withdraw, protocol, decimals) VALUES (2, 'bnb', 2, '0xB8c77482e45F1F44dE1745F52C74426C631bDD52', 0.01, 1, 18);
INSERT INTO public.spot_contracts (id, symbol, chain_id, address, fees_withdraw, protocol, decimals) VALUES (3, 'aave', 2, '0x7fc66500c84a76ad7e9c93437bfc5ac33e2ddae9', 0.35, 1, 18);
INSERT INTO public.spot_contracts (id, symbol, chain_id, address, fees_withdraw, protocol, decimals) VALUES (4, 'btc', 2, '0x0316eb71485b0ab14103307bf65a021042c6d380', 0.002, 1, 18);
INSERT INTO public.spot_contracts (id, symbol, chain_id, address, fees_withdraw, protocol, decimals) VALUES (5, 'omg', 2, '0xd26114cd6EE289AccF82350c8d8487fedB8A0C07', 0.008, 1, 18);
INSERT INTO public.spot_contracts (id, symbol, chain_id, address, fees_withdraw, protocol, decimals) VALUES (6, 'eth', 3, '0x2170ed0880ac9a755fd29b2688956bd959f933f8', 0.006, 6, 18);
INSERT INTO public.spot_contracts (id, symbol, chain_id, address, fees_withdraw, protocol, decimals) VALUES (7, 'tst', 2, '0x31b130CDFcEDB08E6CcA9a0d02964C9a0722D32E', 0.0024, 1, 18);
INSERT INTO public.spot_contracts (id, symbol, chain_id, address, fees_withdraw, protocol, decimals) VALUES (8, 'usdt', 3, '0x55d398326f99059ff775485246999027b3197955', 0.0001, 6, 18);
INSERT INTO public.spot_contracts (id, symbol, chain_id, address, fees_withdraw, protocol, decimals) VALUES (9, 'usdt', 2, '0xdac17f958d2ee523a2206206994597c13d831ec7', 0.001, 1, 18);
INSERT INTO public.spot_contracts (id, symbol, chain_id, address, fees_withdraw, protocol, decimals) VALUES (10, 'usdt', 1, 'TZ7XYSbGhxPKquYp8reVrNn22g7YS58JXa', 6, 9, 8);

INSERT INTO public.spot_currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (1, 'Omisego', 'omg', 0.0100, 100.00000000, 0.0100, 0.0100, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, false, '[2]', true, 0, '2021-12-26 12:27:02.914683+02');
INSERT INTO public.spot_currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (2, 'Binance', 'bnb', 0.0100, 100.00000000, 0.0100, 0.0010, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, true, '[3, 2]', true, 0, '2021-12-26 12:27:02.914683+02');
INSERT INTO public.spot_currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (3, 'Chain Link', 'link', 0.0100, 100.00000000, 0.0100, 0.0100, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, false, '[2]', true, 0, '2021-12-26 12:27:02.914683+02');
INSERT INTO public.spot_currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (4, 'Aave', 'aave', 0.0001, 100.00000000, 0.0001, 0.0001, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, false, '[2]', true, 0, '2021-12-26 12:27:02.914683+02');
INSERT INTO public.spot_currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (5, 'Bitcoin', 'btc', 0.0001, 100.00000000, 0.0001, 0.0001, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, true, '[4, 2]', true, 0, '2021-12-26 12:27:02.914683+02');
INSERT INTO public.spot_currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (6, 'US Dollar', 'usd', 10.0000, 1000.00000000, 5.0000, 1.0000, 100000000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, true, '[7, 6]', true, 1, '2022-08-02 17:18:27.610763+03');
INSERT INTO public.spot_currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (7, 'Euro', 'eur', 0.0001, 100.00000000, 0.0100, 0.0100, 100000000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, true, '[7, 6]', true, 1, '2022-06-11 15:23:00.914358+03');
INSERT INTO public.spot_currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (8, 'Russian ruble', 'rub', 0.0001, 1000000.00000000, 1000.0000, 400.0000, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, false, '[6, 7]', true, 1, '2022-10-10 12:24:01.773841+03');
INSERT INTO public.spot_currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (9, 'Kyrgyzstan Som', 'kgs', 0.0001, 100.00000000, 10.0000, 1.0000, 100000000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, true, '[6, 7]', true, 1, '2022-08-02 16:47:53.279849+03');
INSERT INTO public.spot_currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (10, 'Ukrainian Hryvnia', 'uah', 0.0001, 100.00000000, 0.0100, 0.0100, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, true, '[6, 7]', true, 1, '2022-06-17 17:23:27.806669+03');
INSERT INTO public.spot_currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (11, 'Pound Sterling', 'gbp', 0.0001, 100.00000000, 0.0100, 0.0100, 100000000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, true, '[7, 6]', true, 1, '2022-06-11 15:40:55.332645+03');
INSERT INTO public.spot_currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (12, 'Ethereum', 'eth', 0.0010, 100.00000000, 0.0100, 0.0100, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, true, '[3, 2]', true, 0, '2021-12-26 12:27:02.914683+02');
INSERT INTO public.spot_currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (13, 'Test Token', 'tst', 100.0000, 100000.00000000, 100.0000, 100.0000, 10000000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, false, '[2]', true, 0, '2022-08-07 13:50:25.48916+03');
INSERT INTO public.spot_currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (14, 'Tether USD', 'usdt', 10.0000, 1000.00000000, 5.0000, 1.0000, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, false, '[1, 2, 3]', true, 0, '2021-12-26 12:27:02.914683+02');
INSERT INTO public.spot_currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (15, 'Tron', 'trx', 100.0000, 1000000.00000000, 100.0000, 0.0001, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, true, '[1]', true, 0, '2021-12-26 12:27:02.914683+02');

INSERT INTO public.spot_orders (id, assigning, price, value, quantity, base_unit, quote_unit, user_id, type, status, create_at) VALUES (1, 1, 1254.90731441, 1.9999999000000000, 1.9999999000000000, 'eth', 'usd', 16, 0, 0, '2022-12-08 15:26:16.774524+02');
INSERT INTO public.spot_orders (id, assigning, price, value, quantity, base_unit, quote_unit, user_id, type, status, create_at) VALUES (2, 1, 1254.90731441, 0.4999999900000000, 0.4999999900000000, 'eth', 'usd', 16, 0, 2, '2022-12-08 15:26:27.242877+02');
INSERT INTO public.spot_orders (id, assigning, price, value, quantity, base_unit, quote_unit, user_id, type, status, create_at) VALUES (3, 1, 1254.90731441, 0.3749999900000000, 0.3749999900000000, 'eth', 'usd', 16, 0, 2, '2022-12-08 15:26:29.011507+02');

INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (1, 'eth', 'link', 220.16359184, 6, 4, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (2, 'eth', 'omg', 736.91967577, 6, 4, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (3, 'eth', 'bnb', 5.27048079, 2, 4, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (4, 'trx', 'usd', 0.06819350, 6, 2, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (5, 'btc', 'usd', 22689.08400000, 8, 2, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (6, 'link', 'usd', 7.29913333, 6, 2, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (7, 'omg', 'usd', 2.17870000, 6, 2, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (8, 'bnb', 'usd', 304.17666667, 6, 2, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (9, 'aave', 'usd', 95.36833333, 6, 2, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (10, 'btc', 'bnb', 138.33670553, 8, 4, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (11, 'eth', 'usdt', 1609.71294407, 2, 8, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (12, 'bnb', 'gbp', 189.00000000, 6, 2, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (13, 'bnb', 'trx', 3412.38216665, 6, 8, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (14, 'trx', 'eth', 0.00004233, 6, 6, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (15, 'btc', 'eth', 14.21413631, 8, 4, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (16, 'eth', 'uah', 63020.19688061, 6, 2, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (17, 'eth', 'eur', 1577.38801106, 6, 2, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (18, 'eth', 'gbp', 1329.20534028, 6, 2, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (19, 'bnb', 'uah', 11679.32986432, 6, 2, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (20, 'usdt', 'uah', 39.18388945, 2, 2, false);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (25, 'usd', 'rub', 63.38962512, 2, 4, true);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (26, 'eth', 'usd', 1332.27724865, 6, 2, true);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (27, 'kgs', 'usd', 0.01200000, 2, 6, true);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (21, 'eth', 'tst', 2548.32508030, 8, 4, true);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (22, 'trx', 'usdt', 0.05419917, 2, 8, true);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (23, 'rub', 'kgs', 1.31000000, 2, 8, true);
INSERT INTO public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (24, 'eur', 'rub', 62.42266735, 2, 4, true);

INSERT INTO public.spot_reserves (id, user_id, address, symbol, platform, protocol, value, lock) VALUES (2, 16, '0xc560f3ca3d81a81f3e97464290bcca8c8267643a', 'eth', 1, 0, 2.0000000000000000, false);
INSERT INTO public.spot_reserves (id, user_id, address, symbol, platform, protocol, value, lock) VALUES (1, 16, 'TRHnDwy6qmnZE3PvPdpuZMN7PqRBk28vGe', 'trx', 2, 0, 9001.0000000000000000, false);

INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (12, 'eth', '0x04b59878df90cce3e162e35e672a74b9b708f9ab088294c7bee7ee1bc6a8ef74', 1.0000000000000000, 0, 0, '0xc560f3ca3d81a81f3e97464290bcca8c8267643a', 39, 2, 16, 0, 0, 1, 0, false, 0, 5, '2022-12-02 14:03:57.543933+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (13, 'eth', '0x5052f7b506ef6b7c65dd74d9876df7daade5acb500031840d00de14afdd5a80a', 1.0000000000000000, 0, 0, '0xc560f3ca3d81a81f3e97464290bcca8c8267643a', 44, 2, 16, 0, 0, 1, 0, false, 0, 5, '2022-12-02 14:05:12.729836+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (14, 'eth', '0xd2b17d924f7254438bff62dce42c6f9974b27c901a168b8a8fd65ef2800d09ae', 1.0000000000000000, 0, 0, '0xc560f3ca3d81a81f3e97464290bcca8c8267643a', 182, 2, 16, 0, 0, 1, 0, false, 0, 5, '2022-12-02 14:39:44.899314+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (17, 'trx', 'c1e524ab695079561a9082c48234fafdbb10c1dbdc23ef493507658457d59603', 1000.0000000000000000, 0, 5, 'TRHnDwy6qmnZE3PvPdpuZMN7PqRBk28vGe', 2522, 1, 16, 0, 0, 2, 0, false, 0, 1, '2022-12-02 15:02:13.142671+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (15, 'eth', '0x9170443039346764584a1430ad586225d55dadf2cacb41e7ba36ebe841a87a6a', 1.0000000000000000, 0, 4, '0xc560f3ca3d81a81f3e97464290bcca8c8267643a', 187, 2, 16, 0, 0, 1, 0, false, 0, 1, '2022-12-02 14:40:59.642876+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (16, 'trx', 'b262d5118f50fcfc0635121402cbb2b0e2562368bb1f3597bd54cc9df32a0ac3', 1.0000000000000000, 0, 5, 'TRHnDwy6qmnZE3PvPdpuZMN7PqRBk28vGe', 2516, 1, 16, 0, 0, 2, 0, false, 0, 1, '2022-12-02 15:02:07.005642+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (2, 'trx', 'd44ec74e3c12670e487f75fbcf27e5a54b32bef364605276d3f44d3002d03bf4', 1000.0000000000000000, 0, 5, 'TRHnDwy6qmnZE3PvPdpuZMN7PqRBk28vGe', 119, 1, 16, 0, 0, 2, 0, false, 0, 1, '2022-12-02 12:49:18.132768+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (6, 'trx', 'fd731962689483ca6269fb81e74b80e611776139ccedc25988a7ebbbab544d21', 1000.0000000000000000, 0, 5, 'TRHnDwy6qmnZE3PvPdpuZMN7PqRBk28vGe', 1071, 1, 16, 0, 0, 2, 0, false, 0, 1, '2022-12-02 13:36:54.054519+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (3, 'trx', 'a652528c3d356113348fba644c72e63e9cf63b91424da9feda6300b71750c290', 1000.0000000000000000, 0, 14, 'TRHnDwy6qmnZE3PvPdpuZMN7PqRBk28vGe', 262, 1, 16, 0, 0, 2, 0, false, 0, 1, '2022-12-02 12:56:27.51989+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (5, 'trx', 'b1281b8dced0d818410a0c53397b2322cdca952e4b8f8bd219dd83c07b287527', 1000.0000000000000000, 0, 5, 'TRHnDwy6qmnZE3PvPdpuZMN7PqRBk28vGe', 929, 1, 16, 0, 0, 2, 0, false, 0, 1, '2022-12-02 13:29:48.288926+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (1, 'trx', 'bed22856d59256eae3378c1cf5ab0abd49e3d95a63fd7898d9e72ad59c99e9d2', 1000.0000000000000000, 0, 5, 'TRHnDwy6qmnZE3PvPdpuZMN7PqRBk28vGe', 23, 1, 16, 0, 0, 2, 0, false, 0, 1, '2022-12-02 12:44:30.530807+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (7, 'eth', '0xbe1289b126d17ef5ef2eb7f0000b28d73914af983cc5cdefdaa228fab67bb2e6', 5.0000000000000000, 0, 0, '0xc560f3ca3d81a81f3e97464290bcca8c8267643a', 10, 2, 16, 0, 0, 1, 0, false, 0, 5, '2022-12-02 13:56:42.495368+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (4, 'trx', 'ca1cc77eb60135ae56016f308fb70de5835e621ddf327fcc20c091d7e5b90aab', 1000.0000000000000000, 0, 5, 'TRHnDwy6qmnZE3PvPdpuZMN7PqRBk28vGe', 393, 1, 16, 0, 0, 2, 0, false, 0, 1, '2022-12-02 13:03:00.301542+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (8, 'eth', '0x38639e1909d33bed620a8a84adb3260192cbd2a3ab8610b997f47498fdb0cac1', 5.0000000000000000, 0, 0, '0xc560f3ca3d81a81f3e97464290bcca8c8267643a', 14, 2, 16, 0, 0, 1, 0, false, 0, 5, '2022-12-02 13:57:41.868833+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (9, 'eth', '0x25f9c413838e08793e910406391ee351484cbc77d0363c4ee2ca631140d2eb31', 5.0000000000000000, 0, 0, '0xc560f3ca3d81a81f3e97464290bcca8c8267643a', 17, 2, 16, 0, 0, 1, 0, false, 0, 5, '2022-12-02 13:58:27.059537+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (10, 'eth', '0x95a0c97ff2837e89058f059b815b6b06d4243599423a2e35bb665ee4e0921312', 1.0000000000000000, 0, 0, '0xc560f3ca3d81a81f3e97464290bcca8c8267643a', 26, 2, 16, 0, 0, 1, 0, false, 0, 5, '2022-12-02 14:00:42.517082+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (11, 'eth', '0x07155ba3ef9a499b1808ac689ce18deca21cbd81ccc226a306245520d8c596cc', 1.0000000000000000, 0, 0, '0xc560f3ca3d81a81f3e97464290bcca8c8267643a', 35, 2, 16, 0, 0, 1, 0, false, 0, 5, '2022-12-02 14:02:57.427993+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (18, 'trx', '06b26272e18bb125d3bcf2bf15d4d6c13e1d2b52fca3b2fc97e897de7a6952e4', 1000.0000000000000000, 0, 5, 'TRHnDwy6qmnZE3PvPdpuZMN7PqRBk28vGe', 2765, 1, 16, 0, 0, 2, 0, false, 0, 1, '2022-12-02 15:06:17.484009+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (19, 'eth', '0xe466bb58f935e3b731eb84d4aee0db032144c0f4ee69ac91226d6a22c2b2d744', 1.0000000000000000, 0, 3, '0xc560f3ca3d81a81f3e97464290bcca8c8267643a', 325, 2, 16, 0, 0, 1, 0, false, 0, 1, '2022-12-02 15:15:31.760811+02');
INSERT INTO public.spot_transactions (id, symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol, claim, price, status, create_at) VALUES (20, 'trx', 'be7f0d483357f35e4a3978ef4b3a52352b549473f9c0a8c18434eae3a63ff922', 1000.0000000000000000, 0, 5, 'TRHnDwy6qmnZE3PvPdpuZMN7PqRBk28vGe', 3063, 1, 16, 0, 0, 2, 0, false, 0, 1, '2022-12-02 15:16:31.088611+02');

INSERT INTO public.spot_wallets (id, address, user_id, platform, protocol, symbol) VALUES (1, 'TRHnDwy6qmnZE3PvPdpuZMN7PqRBk28vGe', 16, 2, 0, 'trx');
INSERT INTO public.spot_wallets (id, address, user_id, platform, protocol, symbol) VALUES (2, 'TRHnDwy6qmnZE3PvPdpuZMN7PqRBk28vGe', 16, 2, 9, 'usdt');
INSERT INTO public.spot_wallets (id, address, user_id, platform, protocol, symbol) VALUES (3, '0xc560f3ca3d81a81f3e97464290bcca8c8267643a', 16, 1, 6, 'eth');
INSERT INTO public.spot_wallets (id, address, user_id, platform, protocol, symbol) VALUES (4, '0xc560f3ca3d81a81f3e97464290bcca8c8267643a', 16, 1, 0, 'eth');

SELECT pg_catalog.setval('public.accounts_id_seq', 21, true);
SELECT pg_catalog.setval('public.actions_id_seq', 7, true);
SELECT pg_catalog.setval('public.spot_assets_id_seq', 4, true);
SELECT pg_catalog.setval('public.spot_chains_id_seq', 7, true);
SELECT pg_catalog.setval('public.spot_contracts_id_seq', 10, true);
SELECT pg_catalog.setval('public.spot_currencies_id_seq', 15, true);
SELECT pg_catalog.setval('public.spot_orders_id_seq', 3, true);
SELECT pg_catalog.setval('public.spot_pairs_id_seq', 27, true);
SELECT pg_catalog.setval('public.spot_reserves_id_seq', 2, true);
SELECT pg_catalog.setval('public.spot_trades_id_seq', 17528, true);
SELECT pg_catalog.setval('public.spot_transactions_id_seq', 20, true);
SELECT pg_catalog.setval('public.spot_transfers_id_seq', 1, false);
SELECT pg_catalog.setval('public.spot_wallets_id_seq', 4, true);

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT spot_accounts_pk PRIMARY KEY (id);
ALTER TABLE ONLY public.actions
    ADD CONSTRAINT spot_activities_pk PRIMARY KEY (id);
ALTER TABLE ONLY public.spot_assets
    ADD CONSTRAINT spot_assets_pk PRIMARY KEY (id);
ALTER TABLE ONLY public.spot_chains
    ADD CONSTRAINT spot_chains_pk PRIMARY KEY (id);
ALTER TABLE ONLY public.spot_contracts
    ADD CONSTRAINT spot_contracts_pk PRIMARY KEY (id);
ALTER TABLE ONLY public.spot_currencies
    ADD CONSTRAINT spot_currencies_pk PRIMARY KEY (id);
ALTER TABLE ONLY public.market_orders
    ADD CONSTRAINT market_orders_id_key UNIQUE (id);
ALTER TABLE ONLY public.spot_orders
    ADD CONSTRAINT spot_orders_pk PRIMARY KEY (id);
ALTER TABLE ONLY public.spot_pairs
    ADD CONSTRAINT spot_pairs_pk PRIMARY KEY (id);
ALTER TABLE ONLY public.spot_reserves
    ADD CONSTRAINT spot_reserves_pk PRIMARY KEY (id);
ALTER TABLE ONLY public.spot_transactions
    ADD CONSTRAINT spot_transactions_pk PRIMARY KEY (id);
ALTER TABLE ONLY public.spot_transfers
    ADD CONSTRAINT spot_transfers_pk PRIMARY KEY (id);
ALTER TABLE ONLY public.spot_wallets
    ADD CONSTRAINT spot_wallets_pk PRIMARY KEY (id);

CREATE UNIQUE INDEX accounts_email_uindex ON public.accounts USING btree (email);
CREATE UNIQUE INDEX accounts_id_uindex ON public.accounts USING btree (id);
CREATE UNIQUE INDEX spot_assets_id_uindex ON public.spot_assets USING btree (id);
CREATE UNIQUE INDEX spot_chains_id_uindex ON public.spot_chains USING btree (id);
CREATE UNIQUE INDEX spot_chains_name_uindex ON public.spot_chains USING btree (name);
CREATE UNIQUE INDEX spot_contracts_address_uindex ON public.spot_contracts USING btree (address);
CREATE UNIQUE INDEX spot_contracts_id_uindex ON public.spot_contracts USING btree (id);
CREATE UNIQUE INDEX spot_currencies_id_uindex ON public.spot_currencies USING btree (id);
CREATE UNIQUE INDEX spot_currencies_symbol_uindex ON public.spot_currencies USING btree (symbol);
CREATE UNIQUE INDEX spot_orders_id_uindex ON public.spot_orders USING btree (id);
CREATE UNIQUE INDEX spot_pairs_id_uindex ON public.spot_pairs USING btree (id);
CREATE UNIQUE INDEX spot_reserves_id_uindex ON public.spot_reserves USING btree (id);
CREATE INDEX spot_trades_create_at_idx ON public.spot_trades USING btree (create_at DESC);
CREATE UNIQUE INDEX spot_transactions_id_uindex ON public.spot_transactions USING btree (id);
CREATE UNIQUE INDEX spot_transfers_id_uindex ON public.spot_transfers USING btree (id);
CREATE UNIQUE INDEX spot_wallets_id_uindex ON public.spot_wallets USING btree (id);

GRANT ALL ON DATABASE envoys TO envoys;
