--
-- PostgreSQL database dump
--

-- Dumped from database version 17.5
-- Dumped by pg_dump version 17.5

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: accounts; Type: TABLE; Schema: public; Owner: nelson
--

CREATE TABLE public.accounts (
    id integer NOT NULL,
    name character varying(100) DEFAULT ''::character varying NOT NULL,
    type character varying(100) DEFAULT ''::character varying NOT NULL,
    initial_balance real NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    number character varying(100) NOT NULL
);


ALTER TABLE public.accounts OWNER TO nelson;

--
-- Name: accounts_id_seq; Type: SEQUENCE; Schema: public; Owner: nelson
--

CREATE SEQUENCE public.accounts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.accounts_id_seq OWNER TO nelson;

--
-- Name: accounts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: nelson
--

ALTER SEQUENCE public.accounts_id_seq OWNED BY public.accounts.id;


--
-- Name: categories; Type: TABLE; Schema: public; Owner: nelson
--

CREATE TABLE public.categories (
    id integer NOT NULL,
    name character varying(100) DEFAULT ''::character varying NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    type character varying(50) DEFAULT 'outcome'::character varying NOT NULL
);


ALTER TABLE public.categories OWNER TO nelson;

--
-- Name: categories_id_seq; Type: SEQUENCE; Schema: public; Owner: nelson
--

CREATE SEQUENCE public.categories_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.categories_id_seq OWNER TO nelson;

--
-- Name: categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: nelson
--

ALTER SEQUENCE public.categories_id_seq OWNED BY public.categories.id;


--
-- Name: schema_migration; Type: TABLE; Schema: public; Owner: nelson
--

CREATE TABLE public.schema_migration (
    version character varying(14) NOT NULL
);


ALTER TABLE public.schema_migration OWNER TO nelson;

--
-- Name: transactions; Type: TABLE; Schema: public; Owner: nelson
--

CREATE TABLE public.transactions (
    id integer NOT NULL,
    amount numeric(10,2) NOT NULL,
    description character varying(300) DEFAULT ''::character varying NOT NULL,
    account_id integer NOT NULL,
    category_id integer NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    is_voided boolean DEFAULT false NOT NULL,
    voided_by_transaction_id integer,
    voids_transaction_id integer,
    transaction_date date DEFAULT CURRENT_DATE NOT NULL,
    transaction_number character varying(20) NOT NULL,
    attachment_path text
);


ALTER TABLE public.transactions OWNER TO nelson;

--
-- Name: transactions_id_seq; Type: SEQUENCE; Schema: public; Owner: nelson
--

CREATE SEQUENCE public.transactions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.transactions_id_seq OWNER TO nelson;

--
-- Name: transactions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: nelson
--

ALTER SEQUENCE public.transactions_id_seq OWNED BY public.transactions.id;


--
-- Name: accounts id; Type: DEFAULT; Schema: public; Owner: nelson
--

ALTER TABLE ONLY public.accounts ALTER COLUMN id SET DEFAULT nextval('public.accounts_id_seq'::regclass);


--
-- Name: categories id; Type: DEFAULT; Schema: public; Owner: nelson
--

ALTER TABLE ONLY public.categories ALTER COLUMN id SET DEFAULT nextval('public.categories_id_seq'::regclass);


--
-- Name: transactions id; Type: DEFAULT; Schema: public; Owner: nelson
--

ALTER TABLE ONLY public.transactions ALTER COLUMN id SET DEFAULT nextval('public.transactions_id_seq'::regclass);


--
-- Data for Name: accounts; Type: TABLE DATA; Schema: public; Owner: nelson
--

COPY public.accounts (id, name, type, initial_balance, created_at, updated_at, number) FROM stdin;
4	Cuenta Principal	Ahorros	50000	2025-06-24 15:58:27.749603	2025-06-25 15:14:06.726318	22033844004
12	Cuenta 3	Corriente	2e+06	2025-06-25 12:07:16.728565	2025-07-01 21:39:59.389866	220034345
\.


--
-- Data for Name: categories; Type: TABLE DATA; Schema: public; Owner: nelson
--

COPY public.categories (id, name, created_at, updated_at, type) FROM stdin;
3	Pago de Agua	2025-07-01 12:25:38.896867	2025-07-01 12:25:38.896867	Egreso
5	Anular Transacción E	2025-07-04 08:37:44.079196	2025-07-04 08:37:44.079196	Egreso
6	Anular Transacción I	2025-07-04 08:37:44.079196	2025-07-04 08:37:44.079196	Ingreso
2	Pago de Luz	2025-07-01 11:29:43.451108	2025-07-05 19:19:17.784183	Egreso
7	Tratamiento de Coronas	2025-07-05 19:45:48.916404	2025-07-05 19:45:48.916404	Ingreso
8	Ajuste por Reconciliación	2025-08-01 07:50:32.951299	2025-08-01 07:50:32.951299	Ajuste
\.


--
-- Data for Name: schema_migration; Type: TABLE DATA; Schema: public; Owner: nelson
--

COPY public.schema_migration (version) FROM stdin;
20250612025750
20250612171020
20250612171516
20250612173343
20250612174108
20250612174517
20250624155924
20250629034303
20250704125716
20250704131309
20250704200233
20250706005643
20250706011907
20250718023857
20250801124819
\.


--
-- Data for Name: transactions; Type: TABLE DATA; Schema: public; Owner: nelson
--

COPY public.transactions (id, amount, description, account_id, category_id, created_at, updated_at, is_voided, voided_by_transaction_id, voids_transaction_id, transaction_date, transaction_number, attachment_path) FROM stdin;
15	33.00	dfgd	12	2	2025-07-12 16:24:19.661364	2025-07-12 18:10:07.397473	t	18	\N	2026-11-02	EGR-202611-0002	\N
18	33.00	Anulada por transaccion #EGR-202611-0002: dfgd	12	6	2025-07-16 19:43:37.118301	2025-07-16 19:43:37.118301	f	\N	15	2025-07-16	ANU-202507-0001	\N
17	99.00	hola\t	12	7	2025-07-13 09:35:29.626165	2025-07-13 09:35:29.626165	t	19	\N	2025-08-06	ING-202508-0001	\N
19	99.00	Anulada por transaccion #ING-202508-0001: hola\t	12	5	2025-07-17 18:25:27.934913	2025-07-17 18:25:27.934913	f	\N	17	2025-07-17	ANU-202507-0002	\N
16	200.00	sdfsdf	12	7	2025-07-12 16:51:44.136812	2025-07-12 16:52:56.45305	t	20	\N	2025-01-03	ING-202501-0002	\N
20	200.00	Anulada por transaccion #ING-202501-0002: sdfsdf	12	5	2025-07-17 18:27:09.558887	2025-07-17 18:27:09.558887	f	\N	16	2025-07-17	ANU-202507-0003	\N
13	30.00	Coronas	12	3	2025-07-12 13:23:05.450956	2025-07-12 13:23:32.42407	t	21	\N	2025-01-03	ING-202501-0001	\N
21	30.00	Anulación de la transacción #ING-202501-0001:\nCoronas	12	6	2025-07-17 18:33:51.63936	2025-07-17 18:33:51.63936	f	\N	13	2025-07-17	ANU-202507-0004	\N
12	50.45	Pago de Luz mensual. \naaaaaaaaaaaaaaaaaaaaaaaaaaa\nsdfsdfsdfsdfsdf\nsdfsdfsdf	12	3	2025-07-07 09:17:28.094211	2025-07-08 09:26:06.248346	t	22	\N	2005-01-02	EGR-200501-0001	\N
22	50.45	Anulación de la transacción #EGR-200501-0001:\nPago de Luz mensual. \naaaaaaaaaaaaaaaaaaaaaaaaaaa\nsdfsdfsdfsdfsdf\nsdfsdfsdf	12	6	2025-07-17 18:36:57.9881	2025-07-17 18:36:57.9881	f	\N	12	2025-07-17	ANU-202507-0005	\N
14	333.00	fsdfsd	12	2	2025-07-12 16:23:32.861896	2025-07-12 16:23:32.861896	t	23	\N	2026-11-01	EGR-202611-0001	\N
23	333.00	Anulación de la transacción #EGR-202611-0001:\nfsdfsd	12	6	2025-07-17 18:37:31.754228	2025-07-17 18:37:31.754228	f	\N	14	2025-07-17	ANU-202507-0006	\N
25	222.00	dsfsdf	12	2	2025-07-19 20:41:47.349052	2025-07-19 20:41:47.357688	f	\N	\N	2025-07-19	EGR-202507-0008	tx-25-Arq-Celonis-Stratio.drawio.pdf
24	222.00	con file	12	3	2025-07-18 08:18:17.169044	2025-07-24 09:40:35.369218	f	\N	\N	2025-07-18	EGR-202507-0007	tx-24-2025-01-10-095403_hypr_screenshot.png
\.


--
-- Name: accounts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nelson
--

SELECT pg_catalog.setval('public.accounts_id_seq', 28, true);


--
-- Name: categories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nelson
--

SELECT pg_catalog.setval('public.categories_id_seq', 8, true);


--
-- Name: transactions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nelson
--

SELECT pg_catalog.setval('public.transactions_id_seq', 25, true);


--
-- Name: accounts accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: nelson
--

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: nelson
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: schema_migration schema_migration_pkey; Type: CONSTRAINT; Schema: public; Owner: nelson
--

ALTER TABLE ONLY public.schema_migration
    ADD CONSTRAINT schema_migration_pkey PRIMARY KEY (version);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: nelson
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);


--
-- Name: transactions uq_transaction_number; Type: CONSTRAINT; Schema: public; Owner: nelson
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT uq_transaction_number UNIQUE (transaction_number);


--
-- Name: accounts_name_idx; Type: INDEX; Schema: public; Owner: nelson
--

CREATE UNIQUE INDEX accounts_name_idx ON public.accounts USING btree (name);


--
-- Name: categories_name_idx; Type: INDEX; Schema: public; Owner: nelson
--

CREATE UNIQUE INDEX categories_name_idx ON public.categories USING btree (name);


--
-- Name: schema_migration_version_idx; Type: INDEX; Schema: public; Owner: nelson
--

CREATE UNIQUE INDEX schema_migration_version_idx ON public.schema_migration USING btree (version);


--
-- Name: transactions transactions_account_fk; Type: FK CONSTRAINT; Schema: public; Owner: nelson
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_account_fk FOREIGN KEY (account_id) REFERENCES public.accounts(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: transactions transactions_category_fk; Type: FK CONSTRAINT; Schema: public; Owner: nelson
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_category_fk FOREIGN KEY (category_id) REFERENCES public.categories(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: transactions transactions_voided_by_transaction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: nelson
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_voided_by_transaction_id_fkey FOREIGN KEY (voided_by_transaction_id) REFERENCES public.transactions(id);


--
-- Name: transactions transactions_voids_transaction_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: nelson
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_voids_transaction_id_fkey FOREIGN KEY (voids_transaction_id) REFERENCES public.transactions(id);


--
-- PostgreSQL database dump complete
--

