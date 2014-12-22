--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

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
-- Name: config; Type: TABLE; Schema: public; Owner: operator; Tablespace: 
--

CREATE TABLE config (
    namespace text NOT NULL,
    data json
);


ALTER TABLE public.config OWNER TO operator;

--
-- Name: flavors; Type: TABLE; Schema: public; Owner: operator; Tablespace: 
--

CREATE TABLE flavors (
    flavor_id uuid NOT NULL,
    name text NOT NULL,
    cpu integer NOT NULL,
    memory integer NOT NULL,
    disk integer NOT NULL,
    metadata json DEFAULT '{}'::json NOT NULL
);


ALTER TABLE public.flavors OWNER TO operator;

--
-- Name: hypervisors; Type: TABLE; Schema: public; Owner: operator; Tablespace: 
--

CREATE TABLE hypervisors (
    hypervisor_id uuid NOT NULL,
    mac macaddr NOT NULL,
    ip inet NOT NULL,
    metadata json DEFAULT '{}'::json NOT NULL
);


ALTER TABLE public.hypervisors OWNER TO operator;

--
-- Name: hypervisors_ipranges; Type: TABLE; Schema: public; Owner: operator; Tablespace: 
--

CREATE TABLE hypervisors_ipranges (
    hypervisor_id uuid,
    iprange_id uuid
);


ALTER TABLE public.hypervisors_ipranges OWNER TO operator;

--
-- Name: iprange_networks; Type: TABLE; Schema: public; Owner: operator; Tablespace: 
--

CREATE TABLE iprange_networks (
    iprange_id uuid,
    network_id uuid
);


ALTER TABLE public.iprange_networks OWNER TO operator;

--
-- Name: ipranges; Type: TABLE; Schema: public; Owner: operator; Tablespace: 
--

CREATE TABLE ipranges (
    iprange_id uuid NOT NULL,
    cidr cidr NOT NULL,
    gateway inet NOT NULL,
    start_ip inet NOT NULL,
    end_ip inet NOT NULL,
    metadata json DEFAULT '{}'::json NOT NULL
);


ALTER TABLE public.ipranges OWNER TO operator;

--
-- Name: networks; Type: TABLE; Schema: public; Owner: operator; Tablespace: 
--

CREATE TABLE networks (
    network_id uuid NOT NULL,
    name text NOT NULL,
    metadata json DEFAULT '{}'::json NOT NULL
);


ALTER TABLE public.networks OWNER TO operator;

--
-- Name: permissions; Type: TABLE; Schema: public; Owner: operator; Tablespace: 
--

CREATE TABLE permissions (
    permission_id uuid NOT NULL,
    name text NOT NULL,
    service text NOT NULL,
    action text NOT NULL,
    entitytype text NOT NULL,
    owner boolean DEFAULT true NOT NULL,
    description text,
    metadata json NOT NULL
);


ALTER TABLE public.permissions OWNER TO operator;

--
-- Name: projects; Type: TABLE; Schema: public; Owner: operator; Tablespace: 
--

CREATE TABLE projects (
    project_id uuid NOT NULL,
    name text NOT NULL,
    metadata json DEFAULT '{}'::json NOT NULL
);


ALTER TABLE public.projects OWNER TO operator;

--
-- Name: projects_permissions; Type: TABLE; Schema: public; Owner: operator; Tablespace: 
--

CREATE TABLE projects_permissions (
    project_id uuid,
    permission_id uuid
);


ALTER TABLE public.projects_permissions OWNER TO operator;

--
-- Name: projects_users; Type: TABLE; Schema: public; Owner: operator; Tablespace: 
--

CREATE TABLE projects_users (
    project_id uuid,
    user_id uuid
);


ALTER TABLE public.projects_users OWNER TO operator;

--
-- Name: users; Type: TABLE; Schema: public; Owner: operator; Tablespace: 
--

CREATE TABLE users (
    user_id uuid NOT NULL,
    username text NOT NULL,
    email text,
    metadata json DEFAULT '{}'::json NOT NULL
);


ALTER TABLE public.users OWNER TO operator;

--
-- Name: config_pkey; Type: CONSTRAINT; Schema: public; Owner: operator; Tablespace: 
--

ALTER TABLE ONLY config
    ADD CONSTRAINT config_pkey PRIMARY KEY (namespace);


--
-- Name: flavors_pkey; Type: CONSTRAINT; Schema: public; Owner: operator; Tablespace: 
--

ALTER TABLE ONLY flavors
    ADD CONSTRAINT flavors_pkey PRIMARY KEY (flavor_id);


--
-- Name: hypervisors_pkey; Type: CONSTRAINT; Schema: public; Owner: operator; Tablespace: 
--

ALTER TABLE ONLY hypervisors
    ADD CONSTRAINT hypervisors_pkey PRIMARY KEY (hypervisor_id);


--
-- Name: iprange_networks_iprange_id_key; Type: CONSTRAINT; Schema: public; Owner: operator; Tablespace: 
--

ALTER TABLE ONLY iprange_networks
    ADD CONSTRAINT iprange_networks_iprange_id_key UNIQUE (iprange_id);


--
-- Name: ipranges_pkey; Type: CONSTRAINT; Schema: public; Owner: operator; Tablespace: 
--

ALTER TABLE ONLY ipranges
    ADD CONSTRAINT ipranges_pkey PRIMARY KEY (iprange_id);


--
-- Name: networks_pkey; Type: CONSTRAINT; Schema: public; Owner: operator; Tablespace: 
--

ALTER TABLE ONLY networks
    ADD CONSTRAINT networks_pkey PRIMARY KEY (network_id);


--
-- Name: permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: operator; Tablespace: 
--

ALTER TABLE ONLY permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (permission_id);


--
-- Name: projects_pkey; Type: CONSTRAINT; Schema: public; Owner: operator; Tablespace: 
--

ALTER TABLE ONLY projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (project_id);


--
-- Name: users_pkey; Type: CONSTRAINT; Schema: public; Owner: operator; Tablespace: 
--

ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);


--
-- Name: hypervisors_ipranges_uidx; Type: INDEX; Schema: public; Owner: operator; Tablespace: 
--

CREATE UNIQUE INDEX hypervisors_ipranges_uidx ON hypervisors_ipranges USING btree (hypervisor_id, iprange_id);


--
-- Name: iprange_networks_ukey; Type: INDEX; Schema: public; Owner: operator; Tablespace: 
--

CREATE UNIQUE INDEX iprange_networks_ukey ON iprange_networks USING btree (iprange_id, network_id);


--
-- Name: projects_permissions_uidx; Type: INDEX; Schema: public; Owner: operator; Tablespace: 
--

CREATE UNIQUE INDEX projects_permissions_uidx ON projects_permissions USING btree (project_id, permission_id);


--
-- Name: projects_users_uidx; Type: INDEX; Schema: public; Owner: operator; Tablespace: 
--

CREATE UNIQUE INDEX projects_users_uidx ON projects_users USING btree (project_id, user_id);


--
-- Name: hypervisors_ipranges_hypervisor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: operator
--

ALTER TABLE ONLY hypervisors_ipranges
    ADD CONSTRAINT hypervisors_ipranges_hypervisor_id_fkey FOREIGN KEY (hypervisor_id) REFERENCES hypervisors(hypervisor_id);


--
-- Name: hypervisors_ipranges_iprange_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: operator
--

ALTER TABLE ONLY hypervisors_ipranges
    ADD CONSTRAINT hypervisors_ipranges_iprange_id_fkey FOREIGN KEY (iprange_id) REFERENCES ipranges(iprange_id);


--
-- Name: iprange_networks_iprange_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: operator
--

ALTER TABLE ONLY iprange_networks
    ADD CONSTRAINT iprange_networks_iprange_id_fkey FOREIGN KEY (iprange_id) REFERENCES ipranges(iprange_id);


--
-- Name: iprange_networks_network_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: operator
--

ALTER TABLE ONLY iprange_networks
    ADD CONSTRAINT iprange_networks_network_id_fkey FOREIGN KEY (network_id) REFERENCES networks(network_id);


--
-- Name: projects_permissions_permission_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: operator
--

ALTER TABLE ONLY projects_permissions
    ADD CONSTRAINT projects_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES permissions(permission_id);


--
-- Name: projects_permissions_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: operator
--

ALTER TABLE ONLY projects_permissions
    ADD CONSTRAINT projects_permissions_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(project_id);


--
-- Name: projects_users_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: operator
--

ALTER TABLE ONLY projects_users
    ADD CONSTRAINT projects_users_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(project_id);


--
-- Name: projects_users_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: operator
--

ALTER TABLE ONLY projects_users
    ADD CONSTRAINT projects_users_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(user_id);


--
-- Name: public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM postgres;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- Name: config; Type: ACL; Schema: public; Owner: operator
--

REVOKE ALL ON TABLE config FROM PUBLIC;
REVOKE ALL ON TABLE config FROM operator;
GRANT ALL ON TABLE config TO operator;


--
-- Name: flavors; Type: ACL; Schema: public; Owner: operator
--

REVOKE ALL ON TABLE flavors FROM PUBLIC;
REVOKE ALL ON TABLE flavors FROM operator;
GRANT ALL ON TABLE flavors TO operator;


--
-- Name: hypervisors; Type: ACL; Schema: public; Owner: operator
--

REVOKE ALL ON TABLE hypervisors FROM PUBLIC;
REVOKE ALL ON TABLE hypervisors FROM operator;
GRANT ALL ON TABLE hypervisors TO operator;


--
-- Name: hypervisors_ipranges; Type: ACL; Schema: public; Owner: operator
--

REVOKE ALL ON TABLE hypervisors_ipranges FROM PUBLIC;
REVOKE ALL ON TABLE hypervisors_ipranges FROM operator;
GRANT ALL ON TABLE hypervisors_ipranges TO operator;


--
-- Name: iprange_networks; Type: ACL; Schema: public; Owner: operator
--

REVOKE ALL ON TABLE iprange_networks FROM PUBLIC;
REVOKE ALL ON TABLE iprange_networks FROM operator;
GRANT ALL ON TABLE iprange_networks TO operator;


--
-- Name: ipranges; Type: ACL; Schema: public; Owner: operator
--

REVOKE ALL ON TABLE ipranges FROM PUBLIC;
REVOKE ALL ON TABLE ipranges FROM operator;
GRANT ALL ON TABLE ipranges TO operator;


--
-- Name: networks; Type: ACL; Schema: public; Owner: operator
--

REVOKE ALL ON TABLE networks FROM PUBLIC;
REVOKE ALL ON TABLE networks FROM operator;
GRANT ALL ON TABLE networks TO operator;


--
-- Name: permissions; Type: ACL; Schema: public; Owner: operator
--

REVOKE ALL ON TABLE permissions FROM PUBLIC;
REVOKE ALL ON TABLE permissions FROM operator;
GRANT ALL ON TABLE permissions TO operator;


--
-- Name: projects; Type: ACL; Schema: public; Owner: operator
--

REVOKE ALL ON TABLE projects FROM PUBLIC;
REVOKE ALL ON TABLE projects FROM operator;
GRANT ALL ON TABLE projects TO operator;


--
-- Name: projects_permissions; Type: ACL; Schema: public; Owner: operator
--

REVOKE ALL ON TABLE projects_permissions FROM PUBLIC;
REVOKE ALL ON TABLE projects_permissions FROM operator;
GRANT ALL ON TABLE projects_permissions TO operator;


--
-- Name: projects_users; Type: ACL; Schema: public; Owner: operator
--

REVOKE ALL ON TABLE projects_users FROM PUBLIC;
REVOKE ALL ON TABLE projects_users FROM operator;
GRANT ALL ON TABLE projects_users TO operator;


--
-- Name: users; Type: ACL; Schema: public; Owner: operator
--

REVOKE ALL ON TABLE users FROM PUBLIC;
REVOKE ALL ON TABLE users FROM operator;
GRANT ALL ON TABLE users TO operator;


--
-- PostgreSQL database dump complete
--

