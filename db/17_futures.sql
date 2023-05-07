-- public.futures definition

-- Drop table

-- DROP TABLE public.futures;

CREATE TABLE public.futures (
	id int4 NOT NULL DEFAULT nextval('contracts_id_seq'::regclass),
	assigning varchar(8) NOT NULL DEFAULT 'open'::character varying,
	"position" varchar(8) NOT NULL DEFAULT 'long'::character varying,
	trading varchar(8) NULL,
	base_unit varchar(8) NULL,
	quote_unit varchar(8) NULL,
	price numeric(16, 8) NULL,
	quantity numeric(16, 8) NULL,
	take_profit numeric(16, 8) NULL,
	stop_loss numeric(4, 4) NULL,
	status varchar(8) NULL,
	create_at timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
	leverage numeric(4) NULL DEFAULT 1,
	user_id numeric(8) NOT NULL,
	fees numeric(16, 8) NULL,
	"mode" varchar(8) NULL DEFAULT 'cross'::character varying,
	value numeric(16, 8) NOT NULL DEFAULT 0,
	CONSTRAINT futures_pkey PRIMARY KEY (id)
);

-- Permissions

ALTER TABLE public.futures OWNER TO envoys;
GRANT ALL ON TABLE public.futures TO envoys;