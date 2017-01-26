--
-- PostgreSQL database dump
--

-- Dumped from database version 9.5.5
-- Dumped by pg_dump version 9.5.5

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

SET search_path = public, pg_catalog;

DROP INDEX public.value_idx;
DROP INDEX public.right_key_idx;
DROP INDEX public.left_key_idx;
ALTER TABLE ONLY public.tree DROP CONSTRAINT tree_pkey;
ALTER TABLE ONLY public.goose_db_version DROP CONSTRAINT goose_db_version_pkey;
ALTER TABLE public.tree ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.goose_db_version ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE public.tree_id_seq;
DROP TABLE public.tree;
DROP SEQUENCE public.goose_db_version_id_seq;
DROP TABLE public.goose_db_version;
DROP EXTENSION plpgsql;
DROP SCHEMA public;
--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA public;


--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA public IS 'standard public schema';


--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: goose_db_version; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE goose_db_version (
    id integer NOT NULL,
    version_id bigint NOT NULL,
    is_applied boolean NOT NULL,
    tstamp timestamp without time zone DEFAULT now()
);


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE goose_db_version_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE goose_db_version_id_seq OWNED BY goose_db_version.id;


--
-- Name: tree; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE tree (
    id integer NOT NULL,
    left_key integer,
    right_key integer,
    value character varying(30)
);


--
-- Name: tree_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE tree_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tree_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE tree_id_seq OWNED BY tree.id;


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY goose_db_version ALTER COLUMN id SET DEFAULT nextval('goose_db_version_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY tree ALTER COLUMN id SET DEFAULT nextval('tree_id_seq'::regclass);


--
-- Name: goose_db_version_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);


--
-- Name: tree_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY tree
    ADD CONSTRAINT tree_pkey PRIMARY KEY (id);


--
-- Name: left_key_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX left_key_idx ON tree USING btree (left_key);


--
-- Name: right_key_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX right_key_idx ON tree USING btree (right_key);


--
-- Name: value_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX value_idx ON tree USING btree (value);


--
-- PostgreSQL database dump complete
--

