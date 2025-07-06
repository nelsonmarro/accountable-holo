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
    transaction_date date DEFAULT CURRENT_DATE NOT NULL
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

