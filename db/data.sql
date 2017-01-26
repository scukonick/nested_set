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

--
-- Data for Name: goose_db_version; Type: TABLE DATA; Schema: public; Owner: -
--

COPY goose_db_version (id, version_id, is_applied, tstamp) FROM stdin;
1	0	t	2017-01-25 16:41:02.059897
\.


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('goose_db_version_id_seq', 1, true);


--
-- Data for Name: tree; Type: TABLE DATA; Schema: public; Owner: -
--

COPY tree (id, left_key, right_key, value) FROM stdin;
1	1	18	animals
2	2	7	insects
3	3	4	bees
4	5	6	flies
5	8	13	mammals
6	9	10	dogs
7	11	12	cats
8	14	17	fish
9	15	16	sharks
\.


--
-- Name: tree_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('tree_id_seq', 9, true);


--
-- PostgreSQL database dump complete
--

