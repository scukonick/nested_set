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

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: tree; Type: TABLE; Schema: public; Owner: nest_tree
--

CREATE TABLE tree (
    id integer NOT NULL,
    left_key integer,
    right_key integer,
    value character varying(30)
);


ALTER TABLE tree OWNER TO nest_tree;

--
-- Name: tree_id_seq; Type: SEQUENCE; Schema: public; Owner: nest_tree
--

CREATE SEQUENCE tree_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE tree_id_seq OWNER TO nest_tree;

--
-- Name: tree_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: nest_tree
--

ALTER SEQUENCE tree_id_seq OWNED BY tree.id;


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: nest_tree
--

ALTER TABLE ONLY tree ALTER COLUMN id SET DEFAULT nextval('tree_id_seq'::regclass);


--
-- Data for Name: tree; Type: TABLE DATA; Schema: public; Owner: nest_tree
--

COPY tree (id, left_key, right_key, value) FROM stdin;
5	1	8	animals
6	2	3	dogs
7	4	5	cats
8	6	7	birds
\.


--
-- Name: tree_id_seq; Type: SEQUENCE SET; Schema: public; Owner: nest_tree
--

SELECT pg_catalog.setval('tree_id_seq', 8, true);


--
-- Name: tree_pkey; Type: CONSTRAINT; Schema: public; Owner: nest_tree
--

ALTER TABLE ONLY tree
    ADD CONSTRAINT tree_pkey PRIMARY KEY (id);


--
-- Name: public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM postgres;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

