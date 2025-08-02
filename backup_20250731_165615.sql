--
-- PostgreSQL database dump
--

-- Dumped from database version 15.13
-- Dumped by pg_dump version 15.13

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: cloudbox
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_updated_at_column() OWNER TO cloudbox;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: api_keys; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.api_keys (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone,
    name text NOT NULL,
    key text NOT NULL,
    key_hash text NOT NULL,
    is_active boolean DEFAULT true,
    last_used_at timestamp with time zone,
    expires_at timestamp with time zone,
    permissions text[],
    project_id bigint NOT NULL
);


ALTER TABLE public.api_keys OWNER TO cloudbox;

--
-- Name: api_keys_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.api_keys_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.api_keys_id_seq OWNER TO cloudbox;

--
-- Name: api_keys_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.api_keys_id_seq OWNED BY public.api_keys.id;


--
-- Name: app_sessions; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.app_sessions (
    id character varying(255) NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id text NOT NULL,
    token text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    ip_address text,
    user_agent text,
    device_info jsonb,
    project_id bigint NOT NULL,
    is_active boolean DEFAULT true,
    last_activity timestamp with time zone
);


ALTER TABLE public.app_sessions OWNER TO cloudbox;

--
-- Name: app_users; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.app_users (
    id character varying(255) NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    email text NOT NULL,
    password_hash text NOT NULL,
    name text,
    username text,
    profile_data jsonb,
    preferences jsonb,
    is_active boolean DEFAULT true,
    is_email_verified boolean DEFAULT false,
    last_login_at timestamp with time zone,
    last_seen_at timestamp with time zone,
    project_id bigint NOT NULL,
    login_attempts bigint DEFAULT 0,
    locked_until timestamp with time zone,
    password_reset_token text,
    password_reset_expires timestamp with time zone,
    email_verification_token text
);


ALTER TABLE public.app_users OWNER TO cloudbox;

--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.audit_logs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    action text NOT NULL,
    resource text NOT NULL,
    resource_id text,
    description text,
    actor_id bigint NOT NULL,
    actor_name text NOT NULL,
    actor_role text NOT NULL,
    ip_address text,
    user_agent text,
    method text,
    path text,
    metadata text,
    project_id bigint,
    success boolean DEFAULT true,
    error_msg text
);


ALTER TABLE public.audit_logs OWNER TO cloudbox;

--
-- Name: audit_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.audit_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.audit_logs_id_seq OWNER TO cloudbox;

--
-- Name: audit_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.audit_logs_id_seq OWNED BY public.audit_logs.id;


--
-- Name: backups; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.backups (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone,
    name text NOT NULL,
    description text,
    type text NOT NULL,
    status text DEFAULT 'pending'::character varying,
    size bigint,
    file_path text,
    checksum text,
    completed_at timestamp with time zone,
    project_id bigint NOT NULL
);


ALTER TABLE public.backups OWNER TO cloudbox;

--
-- Name: backups_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.backups_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.backups_id_seq OWNER TO cloudbox;

--
-- Name: backups_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.backups_id_seq OWNED BY public.backups.id;


--
-- Name: buckets; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.buckets (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    description text,
    max_file_size bigint DEFAULT 52428800,
    allowed_types jsonb,
    is_public boolean DEFAULT false,
    project_id bigint NOT NULL,
    file_count bigint DEFAULT 0,
    total_size bigint DEFAULT 0,
    last_modified timestamp with time zone
);


ALTER TABLE public.buckets OWNER TO cloudbox;

--
-- Name: buckets_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.buckets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.buckets_id_seq OWNER TO cloudbox;

--
-- Name: buckets_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.buckets_id_seq OWNED BY public.buckets.id;


--
-- Name: channel_members; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.channel_members (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    channel_id text NOT NULL,
    user_id text NOT NULL,
    role text DEFAULT 'member'::text,
    project_id bigint NOT NULL,
    is_active boolean DEFAULT true,
    joined_at timestamp with time zone,
    last_read_at timestamp with time zone,
    is_muted boolean DEFAULT false,
    can_read boolean DEFAULT true,
    can_write boolean DEFAULT true,
    can_invite boolean DEFAULT false,
    can_moderate boolean DEFAULT false
);


ALTER TABLE public.channel_members OWNER TO cloudbox;

--
-- Name: channel_members_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.channel_members_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.channel_members_id_seq OWNER TO cloudbox;

--
-- Name: channel_members_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.channel_members_id_seq OWNED BY public.channel_members.id;


--
-- Name: channels; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.channels (
    id character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone,
    name text NOT NULL,
    description text,
    type text DEFAULT 'public'::character varying NOT NULL,
    is_active boolean DEFAULT true,
    settings jsonb DEFAULT '{}'::jsonb,
    last_activity timestamp with time zone DEFAULT now(),
    message_count bigint DEFAULT 0,
    member_count bigint DEFAULT 0,
    project_id bigint NOT NULL,
    created_by text DEFAULT 'system'::character varying,
    topic text,
    max_members bigint DEFAULT 0
);


ALTER TABLE public.channels OWNER TO cloudbox;

--
-- Name: collections; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.collections (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    description text,
    schema jsonb,
    indexes jsonb,
    project_id bigint NOT NULL,
    document_count bigint DEFAULT 0,
    last_modified timestamp with time zone
);


ALTER TABLE public.collections OWNER TO cloudbox;

--
-- Name: collections_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.collections_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.collections_id_seq OWNER TO cloudbox;

--
-- Name: collections_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.collections_id_seq OWNED BY public.collections.id;


--
-- Name: cors_configs; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.cors_configs (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone,
    allowed_origins text[],
    allowed_methods text[],
    allowed_headers text[],
    exposed_headers text[],
    allow_credentials boolean DEFAULT false,
    max_age bigint DEFAULT 3600,
    project_id bigint NOT NULL
);


ALTER TABLE public.cors_configs OWNER TO cloudbox;

--
-- Name: cors_configs_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.cors_configs_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.cors_configs_id_seq OWNER TO cloudbox;

--
-- Name: cors_configs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.cors_configs_id_seq OWNED BY public.cors_configs.id;


--
-- Name: deployments; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.deployments (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone,
    version text NOT NULL,
    status text DEFAULT 'pending'::character varying,
    build_logs text,
    deployed_at timestamp with time zone,
    domain text,
    environment jsonb,
    file_count bigint,
    total_size bigint,
    build_time bigint,
    project_id bigint NOT NULL,
    name text NOT NULL,
    description text,
    github_repository_id integer,
    web_server_id bigint NOT NULL,
    port bigint DEFAULT 3000,
    branch text DEFAULT 'main'::character varying,
    commit_hash text,
    build_command text,
    start_command text,
    deploy_logs text,
    error_logs text,
    deploy_time bigint,
    is_auto_deploy_enabled boolean DEFAULT false,
    trigger_branch text DEFAULT 'main'::text,
    subdomain text,
    commit_message text,
    commit_author text,
    git_hub_repository_id bigint NOT NULL
);


ALTER TABLE public.deployments OWNER TO cloudbox;

--
-- Name: deployments_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.deployments_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.deployments_id_seq OWNER TO cloudbox;

--
-- Name: deployments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.deployments_id_seq OWNED BY public.deployments.id;


--
-- Name: documents; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.documents (
    id character varying(255) NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    collection_name text NOT NULL,
    project_id bigint NOT NULL,
    data jsonb,
    version bigint DEFAULT 1,
    author text
);


ALTER TABLE public.documents OWNER TO cloudbox;

--
-- Name: files; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.files (
    id character varying(255) NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    original_name text NOT NULL,
    file_name text NOT NULL,
    file_path text NOT NULL,
    mime_type text NOT NULL,
    size bigint NOT NULL,
    checksum text,
    bucket_name text NOT NULL,
    folder_path text,
    project_id bigint NOT NULL,
    is_public boolean DEFAULT false,
    author text,
    public_url text,
    private_url text
);


ALTER TABLE public.files OWNER TO cloudbox;

--
-- Name: function_domains; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.function_domains (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    domain text NOT NULL,
    is_verified boolean DEFAULT false,
    certificate text,
    function_id bigint NOT NULL,
    project_id bigint NOT NULL
);


ALTER TABLE public.function_domains OWNER TO cloudbox;

--
-- Name: function_domains_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.function_domains_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.function_domains_id_seq OWNER TO cloudbox;

--
-- Name: function_domains_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.function_domains_id_seq OWNED BY public.function_domains.id;


--
-- Name: function_executions; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.function_executions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    function_id bigint NOT NULL,
    execution_id text NOT NULL,
    request_data jsonb,
    response_data jsonb,
    headers jsonb,
    method text NOT NULL,
    path text,
    status text NOT NULL,
    status_code bigint DEFAULT 200,
    execution_time bigint,
    memory_usage bigint,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    logs text,
    error_message text,
    user_agent text,
    client_ip text,
    source text DEFAULT 'http'::text,
    project_id bigint NOT NULL
);


ALTER TABLE public.function_executions OWNER TO cloudbox;

--
-- Name: function_executions_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.function_executions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.function_executions_id_seq OWNER TO cloudbox;

--
-- Name: function_executions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.function_executions_id_seq OWNED BY public.function_executions.id;


--
-- Name: functions; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.functions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    description text,
    runtime text DEFAULT 'nodejs18'::text NOT NULL,
    language text DEFAULT 'javascript'::text NOT NULL,
    code text NOT NULL,
    entry_point text DEFAULT 'index.handler'::text,
    timeout bigint DEFAULT 30,
    memory bigint DEFAULT 128,
    environment jsonb,
    commands jsonb,
    dependencies jsonb,
    status text DEFAULT 'draft'::text,
    version bigint DEFAULT 1,
    last_deployed_at timestamp with time zone,
    build_logs text,
    deployment_logs text,
    error_message text,
    function_url text,
    is_active boolean DEFAULT true,
    is_public boolean DEFAULT false,
    project_id bigint NOT NULL
);


ALTER TABLE public.functions OWNER TO cloudbox;

--
-- Name: functions_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.functions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.functions_id_seq OWNER TO cloudbox;

--
-- Name: functions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.functions_id_seq OWNED BY public.functions.id;


--
-- Name: git_hub_repositories; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.git_hub_repositories (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone,
    name text NOT NULL,
    full_name text NOT NULL,
    clone_url text NOT NULL,
    description text,
    branch text DEFAULT 'main'::character varying,
    is_private boolean DEFAULT false,
    build_command text DEFAULT 'npm run build'::text,
    start_command text DEFAULT 'npm start'::text,
    app_port bigint DEFAULT 3000,
    is_active boolean DEFAULT true,
    last_sync_at timestamp with time zone,
    last_commit_hash text DEFAULT ''::character varying,
    pending_commit_hash text DEFAULT ''::character varying,
    pending_commit_branch text DEFAULT ''::character varying,
    has_pending_update boolean DEFAULT false,
    project_id bigint NOT NULL,
    webhook_id bigint,
    webhook_secret text,
    ssh_key_id bigint,
    sdk_version text,
    environment jsonb,
    git_hub_id bigint,
    default_branch text,
    language text,
    size bigint,
    stargazers_count bigint,
    forks_count bigint
);


ALTER TABLE public.git_hub_repositories OWNER TO cloudbox;

--
-- Name: git_hub_repositories_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.git_hub_repositories_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.git_hub_repositories_id_seq OWNER TO cloudbox;

--
-- Name: git_hub_repositories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.git_hub_repositories_id_seq OWNED BY public.git_hub_repositories.id;


--
-- Name: host_key_entries; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.host_key_entries (
    id bigint NOT NULL,
    hostname text NOT NULL,
    port bigint DEFAULT 22 NOT NULL,
    key_type text NOT NULL,
    public_key text NOT NULL,
    fingerprint text NOT NULL,
    first_seen bigint NOT NULL,
    last_seen bigint NOT NULL,
    verified boolean DEFAULT false,
    project_id bigint NOT NULL
);


ALTER TABLE public.host_key_entries OWNER TO cloudbox;

--
-- Name: host_key_entries_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.host_key_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.host_key_entries_id_seq OWNER TO cloudbox;

--
-- Name: host_key_entries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.host_key_entries_id_seq OWNED BY public.host_key_entries.id;


--
-- Name: message_reactions; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.message_reactions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    message_id text NOT NULL,
    user_id text NOT NULL,
    emoji text NOT NULL,
    project_id bigint NOT NULL
);


ALTER TABLE public.message_reactions OWNER TO cloudbox;

--
-- Name: message_reactions_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.message_reactions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.message_reactions_id_seq OWNER TO cloudbox;

--
-- Name: message_reactions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.message_reactions_id_seq OWNED BY public.message_reactions.id;


--
-- Name: message_reads; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.message_reads (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    message_id text NOT NULL,
    user_id text NOT NULL,
    read_at timestamp with time zone NOT NULL,
    project_id bigint NOT NULL
);


ALTER TABLE public.message_reads OWNER TO cloudbox;

--
-- Name: message_reads_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.message_reads_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.message_reads_id_seq OWNER TO cloudbox;

--
-- Name: message_reads_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.message_reads_id_seq OWNED BY public.message_reads.id;


--
-- Name: messages; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.messages (
    id character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone,
    content text NOT NULL,
    type text DEFAULT 'text'::text,
    metadata jsonb DEFAULT '{}'::jsonb,
    channel_id text NOT NULL,
    user_id text NOT NULL,
    project_id bigint NOT NULL,
    parent_id text,
    thread_id text,
    reply_count bigint DEFAULT 0,
    is_edited boolean DEFAULT false,
    edited_at timestamp with time zone,
    is_deleted boolean DEFAULT false,
    message_deleted_at timestamp with time zone,
    reaction_count bigint DEFAULT 0
);


ALTER TABLE public.messages OWNER TO cloudbox;

--
-- Name: organizations; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.organizations (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    description text,
    color text DEFAULT '#3B82F6'::text,
    is_active boolean DEFAULT true,
    website text,
    email text,
    phone text,
    contact_person text,
    logo_url text,
    logo_file_id bigint,
    address text,
    city text,
    country text,
    postal_code text,
    user_id bigint NOT NULL,
    project_count bigint DEFAULT 0
);


ALTER TABLE public.organizations OWNER TO cloudbox;

--
-- Name: organizations_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.organizations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.organizations_id_seq OWNER TO cloudbox;

--
-- Name: organizations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.organizations_id_seq OWNED BY public.organizations.id;


--
-- Name: projects; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.projects (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone,
    name text NOT NULL,
    description text,
    slug text NOT NULL,
    is_active boolean DEFAULT true,
    user_id bigint NOT NULL,
    organization_id bigint
);


ALTER TABLE public.projects OWNER TO cloudbox;

--
-- Name: projects_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.projects_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.projects_id_seq OWNER TO cloudbox;

--
-- Name: projects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.projects_id_seq OWNED BY public.projects.id;


--
-- Name: refresh_tokens; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.refresh_tokens (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    token text NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    is_active boolean DEFAULT true,
    ip_address text,
    user_agent text,
    user_id bigint NOT NULL
);


ALTER TABLE public.refresh_tokens OWNER TO cloudbox;

--
-- Name: refresh_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.refresh_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.refresh_tokens_id_seq OWNER TO cloudbox;

--
-- Name: refresh_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.refresh_tokens_id_seq OWNED BY public.refresh_tokens.id;


--
-- Name: ssh_keys; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.ssh_keys (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone,
    name text NOT NULL,
    public_key text NOT NULL,
    private_key text NOT NULL,
    fingerprint text NOT NULL,
    is_active boolean DEFAULT true,
    project_id bigint NOT NULL,
    description text,
    key_type text DEFAULT 'rsa'::text,
    key_size bigint DEFAULT 2048,
    last_used_at timestamp with time zone,
    server_count bigint DEFAULT 0
);


ALTER TABLE public.ssh_keys OWNER TO cloudbox;

--
-- Name: ssh_keys_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.ssh_keys_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ssh_keys_id_seq OWNER TO cloudbox;

--
-- Name: ssh_keys_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.ssh_keys_id_seq OWNED BY public.ssh_keys.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.users (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone,
    email text NOT NULL,
    password_hash text NOT NULL,
    name text,
    role character varying(20) DEFAULT 'admin'::character varying,
    is_active boolean DEFAULT true,
    last_login_at timestamp with time zone
);


ALTER TABLE public.users OWNER TO cloudbox;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_id_seq OWNER TO cloudbox;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: web_servers; Type: TABLE; Schema: public; Owner: cloudbox
--

CREATE TABLE public.web_servers (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone,
    name text NOT NULL,
    hostname text NOT NULL,
    port bigint DEFAULT 22,
    username text NOT NULL,
    is_active boolean DEFAULT true,
    last_ping_at timestamp with time zone,
    project_id bigint NOT NULL,
    ssh_key_id bigint NOT NULL,
    description text,
    server_type text DEFAULT 'vps'::text,
    os text DEFAULT 'ubuntu'::text,
    docker_enabled boolean DEFAULT true,
    nginx_enabled boolean DEFAULT true,
    deploy_path text DEFAULT '/var/www'::text,
    backup_path text DEFAULT '/var/backups'::text,
    log_path text DEFAULT '/var/log/deployments'::text,
    last_connected_at timestamp with time zone,
    connection_status text DEFAULT 'unknown'::text
);


ALTER TABLE public.web_servers OWNER TO cloudbox;

--
-- Name: web_servers_id_seq; Type: SEQUENCE; Schema: public; Owner: cloudbox
--

CREATE SEQUENCE public.web_servers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.web_servers_id_seq OWNER TO cloudbox;

--
-- Name: web_servers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloudbox
--

ALTER SEQUENCE public.web_servers_id_seq OWNED BY public.web_servers.id;


--
-- Name: api_keys id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.api_keys ALTER COLUMN id SET DEFAULT nextval('public.api_keys_id_seq'::regclass);


--
-- Name: audit_logs id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.audit_logs ALTER COLUMN id SET DEFAULT nextval('public.audit_logs_id_seq'::regclass);


--
-- Name: backups id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.backups ALTER COLUMN id SET DEFAULT nextval('public.backups_id_seq'::regclass);


--
-- Name: buckets id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.buckets ALTER COLUMN id SET DEFAULT nextval('public.buckets_id_seq'::regclass);


--
-- Name: channel_members id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.channel_members ALTER COLUMN id SET DEFAULT nextval('public.channel_members_id_seq'::regclass);


--
-- Name: collections id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.collections ALTER COLUMN id SET DEFAULT nextval('public.collections_id_seq'::regclass);


--
-- Name: cors_configs id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.cors_configs ALTER COLUMN id SET DEFAULT nextval('public.cors_configs_id_seq'::regclass);


--
-- Name: deployments id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.deployments ALTER COLUMN id SET DEFAULT nextval('public.deployments_id_seq'::regclass);


--
-- Name: function_domains id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.function_domains ALTER COLUMN id SET DEFAULT nextval('public.function_domains_id_seq'::regclass);


--
-- Name: function_executions id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.function_executions ALTER COLUMN id SET DEFAULT nextval('public.function_executions_id_seq'::regclass);


--
-- Name: functions id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.functions ALTER COLUMN id SET DEFAULT nextval('public.functions_id_seq'::regclass);


--
-- Name: git_hub_repositories id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.git_hub_repositories ALTER COLUMN id SET DEFAULT nextval('public.git_hub_repositories_id_seq'::regclass);


--
-- Name: host_key_entries id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.host_key_entries ALTER COLUMN id SET DEFAULT nextval('public.host_key_entries_id_seq'::regclass);


--
-- Name: message_reactions id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.message_reactions ALTER COLUMN id SET DEFAULT nextval('public.message_reactions_id_seq'::regclass);


--
-- Name: message_reads id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.message_reads ALTER COLUMN id SET DEFAULT nextval('public.message_reads_id_seq'::regclass);


--
-- Name: organizations id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.organizations ALTER COLUMN id SET DEFAULT nextval('public.organizations_id_seq'::regclass);


--
-- Name: projects id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.projects ALTER COLUMN id SET DEFAULT nextval('public.projects_id_seq'::regclass);


--
-- Name: refresh_tokens id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.refresh_tokens ALTER COLUMN id SET DEFAULT nextval('public.refresh_tokens_id_seq'::regclass);


--
-- Name: ssh_keys id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.ssh_keys ALTER COLUMN id SET DEFAULT nextval('public.ssh_keys_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: web_servers id; Type: DEFAULT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.web_servers ALTER COLUMN id SET DEFAULT nextval('public.web_servers_id_seq'::regclass);


--
-- Data for Name: api_keys; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.api_keys (id, created_at, updated_at, deleted_at, name, key, key_hash, is_active, last_used_at, expires_at, permissions, project_id) FROM stdin;
1	2025-07-31 13:43:35.523039+00	2025-07-31 13:43:35.523039+00	\N	Default API Key	cbx_demo_key_12345678901234567890	$2a$10$demo_key_hash_placeholder_value_here	t	\N	\N	{read,write,admin}	1
\.


--
-- Data for Name: app_sessions; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.app_sessions (id, created_at, updated_at, deleted_at, user_id, token, expires_at, ip_address, user_agent, device_info, project_id, is_active, last_activity) FROM stdin;
\.


--
-- Data for Name: app_users; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.app_users (id, created_at, updated_at, deleted_at, email, password_hash, name, username, profile_data, preferences, is_active, is_email_verified, last_login_at, last_seen_at, project_id, login_attempts, locked_until, password_reset_token, password_reset_expires, email_verification_token) FROM stdin;
\.


--
-- Data for Name: audit_logs; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.audit_logs (id, created_at, updated_at, deleted_at, action, resource, resource_id, description, actor_id, actor_name, actor_role, ip_address, user_agent, method, path, metadata, project_id, success, error_msg) FROM stdin;
\.


--
-- Data for Name: backups; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.backups (id, created_at, updated_at, deleted_at, name, description, type, status, size, file_path, checksum, completed_at, project_id) FROM stdin;
\.


--
-- Data for Name: buckets; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.buckets (id, created_at, updated_at, deleted_at, name, description, max_file_size, allowed_types, is_public, project_id, file_count, total_size, last_modified) FROM stdin;
\.


--
-- Data for Name: channel_members; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.channel_members (id, created_at, updated_at, deleted_at, channel_id, user_id, role, project_id, is_active, joined_at, last_read_at, is_muted, can_read, can_write, can_invite, can_moderate) FROM stdin;
\.


--
-- Data for Name: channels; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.channels (id, created_at, updated_at, deleted_at, name, description, type, is_active, settings, last_activity, message_count, member_count, project_id, created_by, topic, max_members) FROM stdin;
\.


--
-- Data for Name: collections; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.collections (id, created_at, updated_at, deleted_at, name, description, schema, indexes, project_id, document_count, last_modified) FROM stdin;
\.


--
-- Data for Name: cors_configs; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.cors_configs (id, created_at, updated_at, deleted_at, allowed_origins, allowed_methods, allowed_headers, exposed_headers, allow_credentials, max_age, project_id) FROM stdin;
\.


--
-- Data for Name: deployments; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.deployments (id, created_at, updated_at, deleted_at, version, status, build_logs, deployed_at, domain, environment, file_count, total_size, build_time, project_id, name, description, github_repository_id, web_server_id, port, branch, commit_hash, build_command, start_command, deploy_logs, error_logs, deploy_time, is_auto_deploy_enabled, trigger_branch, subdomain, commit_message, commit_author, git_hub_repository_id) FROM stdin;
\.


--
-- Data for Name: documents; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.documents (id, created_at, updated_at, deleted_at, collection_name, project_id, data, version, author) FROM stdin;
\.


--
-- Data for Name: files; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.files (id, created_at, updated_at, deleted_at, original_name, file_name, file_path, mime_type, size, checksum, bucket_name, folder_path, project_id, is_public, author, public_url, private_url) FROM stdin;
\.


--
-- Data for Name: function_domains; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.function_domains (id, created_at, updated_at, deleted_at, domain, is_verified, certificate, function_id, project_id) FROM stdin;
\.


--
-- Data for Name: function_executions; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.function_executions (id, created_at, updated_at, deleted_at, function_id, execution_id, request_data, response_data, headers, method, path, status, status_code, execution_time, memory_usage, started_at, completed_at, logs, error_message, user_agent, client_ip, source, project_id) FROM stdin;
\.


--
-- Data for Name: functions; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.functions (id, created_at, updated_at, deleted_at, name, description, runtime, language, code, entry_point, timeout, memory, environment, commands, dependencies, status, version, last_deployed_at, build_logs, deployment_logs, error_message, function_url, is_active, is_public, project_id) FROM stdin;
\.


--
-- Data for Name: git_hub_repositories; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.git_hub_repositories (id, created_at, updated_at, deleted_at, name, full_name, clone_url, description, branch, is_private, build_command, start_command, app_port, is_active, last_sync_at, last_commit_hash, pending_commit_hash, pending_commit_branch, has_pending_update, project_id, webhook_id, webhook_secret, ssh_key_id, sdk_version, environment, git_hub_id, default_branch, language, size, stargazers_count, forks_count) FROM stdin;
1	2025-07-31 13:43:35.524266+00	2025-07-31 13:43:35.524266+00	\N	sample-app	user/sample-app	https://github.com/user/sample-app.git	Sample application for CloudBox testing	main	f	npm run build	npm start	3000	t	\N				f	1	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N
\.


--
-- Data for Name: host_key_entries; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.host_key_entries (id, hostname, port, key_type, public_key, fingerprint, first_seen, last_seen, verified, project_id) FROM stdin;
\.


--
-- Data for Name: message_reactions; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.message_reactions (id, created_at, updated_at, deleted_at, message_id, user_id, emoji, project_id) FROM stdin;
\.


--
-- Data for Name: message_reads; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.message_reads (id, created_at, updated_at, deleted_at, message_id, user_id, read_at, project_id) FROM stdin;
\.


--
-- Data for Name: messages; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.messages (id, created_at, updated_at, deleted_at, content, type, metadata, channel_id, user_id, project_id, parent_id, thread_id, reply_count, is_edited, edited_at, is_deleted, message_deleted_at, reaction_count) FROM stdin;
\.


--
-- Data for Name: organizations; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.organizations (id, created_at, updated_at, deleted_at, name, description, color, is_active, website, email, phone, contact_person, logo_url, logo_file_id, address, city, country, postal_code, user_id, project_count) FROM stdin;
\.


--
-- Data for Name: projects; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.projects (id, created_at, updated_at, deleted_at, name, description, slug, is_active, user_id, organization_id) FROM stdin;
1	2025-07-31 13:43:35.522168+00	2025-07-31 13:43:35.522168+00	\N	Demo Project	Default demo project for CloudBox	demo-project	t	1	\N
\.


--
-- Data for Name: refresh_tokens; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.refresh_tokens (id, created_at, updated_at, deleted_at, token, token_hash, expires_at, is_active, ip_address, user_agent, user_id) FROM stdin;
\.


--
-- Data for Name: ssh_keys; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.ssh_keys (id, created_at, updated_at, deleted_at, name, public_key, private_key, fingerprint, is_active, project_id, description, key_type, key_size, last_used_at, server_count) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.users (id, created_at, updated_at, deleted_at, email, password_hash, name, role, is_active, last_login_at) FROM stdin;
2	2025-07-31 13:47:55.730253+00	2025-07-31 13:47:55.730253+00	\N	admin@cloudbox.local	\\$2a\\$10\\$72Eg6eu/TToC/T5MzsnEuOwbmp8ITu0m1LfYiDY3KmGofxkwEZCD.	CloudBox Admin	superadmin	t	\N
1	2025-07-31 13:43:35.521549+00	2025-07-31 14:32:13.651896+00	\N	admin@cloudbox.dev	$2b$12$XuhuQlzkg3wACdXfAU9mu.NMMtcvhwaf91N2S9saHt3mWjDArdAcG	Admin User	superadmin	t	2025-07-31 14:32:13.651774+00
\.


--
-- Data for Name: web_servers; Type: TABLE DATA; Schema: public; Owner: cloudbox
--

COPY public.web_servers (id, created_at, updated_at, deleted_at, name, hostname, port, username, is_active, last_ping_at, project_id, ssh_key_id, description, server_type, os, docker_enabled, nginx_enabled, deploy_path, backup_path, log_path, last_connected_at, connection_status) FROM stdin;
\.


--
-- Name: api_keys_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.api_keys_id_seq', 1, true);


--
-- Name: audit_logs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.audit_logs_id_seq', 1, false);


--
-- Name: backups_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.backups_id_seq', 1, false);


--
-- Name: buckets_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.buckets_id_seq', 1, false);


--
-- Name: channel_members_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.channel_members_id_seq', 1, false);


--
-- Name: collections_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.collections_id_seq', 1, false);


--
-- Name: cors_configs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.cors_configs_id_seq', 1, false);


--
-- Name: deployments_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.deployments_id_seq', 1, false);


--
-- Name: function_domains_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.function_domains_id_seq', 1, false);


--
-- Name: function_executions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.function_executions_id_seq', 1, false);


--
-- Name: functions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.functions_id_seq', 1, false);


--
-- Name: git_hub_repositories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.git_hub_repositories_id_seq', 1, true);


--
-- Name: host_key_entries_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.host_key_entries_id_seq', 1, false);


--
-- Name: message_reactions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.message_reactions_id_seq', 1, false);


--
-- Name: message_reads_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.message_reads_id_seq', 1, false);


--
-- Name: organizations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.organizations_id_seq', 1, false);


--
-- Name: projects_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.projects_id_seq', 1, true);


--
-- Name: refresh_tokens_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.refresh_tokens_id_seq', 1, false);


--
-- Name: ssh_keys_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.ssh_keys_id_seq', 1, false);


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.users_id_seq', 2, true);


--
-- Name: web_servers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloudbox
--

SELECT pg_catalog.setval('public.web_servers_id_seq', 1, false);


--
-- Name: api_keys api_keys_key_key; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_key_key UNIQUE (key);


--
-- Name: api_keys api_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);


--
-- Name: app_sessions app_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.app_sessions
    ADD CONSTRAINT app_sessions_pkey PRIMARY KEY (id);


--
-- Name: app_users app_users_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.app_users
    ADD CONSTRAINT app_users_pkey PRIMARY KEY (id);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: backups backups_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.backups
    ADD CONSTRAINT backups_pkey PRIMARY KEY (id);


--
-- Name: buckets buckets_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.buckets
    ADD CONSTRAINT buckets_pkey PRIMARY KEY (id);


--
-- Name: channel_members channel_members_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.channel_members
    ADD CONSTRAINT channel_members_pkey PRIMARY KEY (id);


--
-- Name: channels channels_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.channels
    ADD CONSTRAINT channels_pkey PRIMARY KEY (id);


--
-- Name: collections collections_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.collections
    ADD CONSTRAINT collections_pkey PRIMARY KEY (id);


--
-- Name: cors_configs cors_configs_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.cors_configs
    ADD CONSTRAINT cors_configs_pkey PRIMARY KEY (id);


--
-- Name: cors_configs cors_configs_project_id_key; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.cors_configs
    ADD CONSTRAINT cors_configs_project_id_key UNIQUE (project_id);


--
-- Name: deployments deployments_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.deployments
    ADD CONSTRAINT deployments_pkey PRIMARY KEY (id);


--
-- Name: documents documents_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_pkey PRIMARY KEY (id);


--
-- Name: files files_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_pkey PRIMARY KEY (id);


--
-- Name: function_domains function_domains_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.function_domains
    ADD CONSTRAINT function_domains_pkey PRIMARY KEY (id);


--
-- Name: function_executions function_executions_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.function_executions
    ADD CONSTRAINT function_executions_pkey PRIMARY KEY (id);


--
-- Name: functions functions_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.functions
    ADD CONSTRAINT functions_pkey PRIMARY KEY (id);


--
-- Name: git_hub_repositories git_hub_repositories_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.git_hub_repositories
    ADD CONSTRAINT git_hub_repositories_pkey PRIMARY KEY (id);


--
-- Name: host_key_entries host_key_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.host_key_entries
    ADD CONSTRAINT host_key_entries_pkey PRIMARY KEY (id);


--
-- Name: message_reactions message_reactions_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.message_reactions
    ADD CONSTRAINT message_reactions_pkey PRIMARY KEY (id);


--
-- Name: message_reads message_reads_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.message_reads
    ADD CONSTRAINT message_reads_pkey PRIMARY KEY (id);


--
-- Name: messages messages_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id);


--
-- Name: organizations organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);


--
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);


--
-- Name: projects projects_slug_key; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_slug_key UNIQUE (slug);


--
-- Name: refresh_tokens refresh_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_pkey PRIMARY KEY (id);


--
-- Name: ssh_keys ssh_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.ssh_keys
    ADD CONSTRAINT ssh_keys_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: web_servers web_servers_pkey; Type: CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.web_servers
    ADD CONSTRAINT web_servers_pkey PRIMARY KEY (id);


--
-- Name: idx_api_keys_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_api_keys_deleted_at ON public.api_keys USING btree (deleted_at);


--
-- Name: idx_api_keys_key; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_api_keys_key ON public.api_keys USING btree (key);


--
-- Name: idx_api_keys_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_api_keys_project_id ON public.api_keys USING btree (project_id);


--
-- Name: idx_app_sessions_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_app_sessions_deleted_at ON public.app_sessions USING btree (deleted_at);


--
-- Name: idx_app_sessions_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_app_sessions_project_id ON public.app_sessions USING btree (project_id);


--
-- Name: idx_app_sessions_token; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE UNIQUE INDEX idx_app_sessions_token ON public.app_sessions USING btree (token);


--
-- Name: idx_app_sessions_user_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_app_sessions_user_id ON public.app_sessions USING btree (user_id);


--
-- Name: idx_app_users_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_app_users_deleted_at ON public.app_users USING btree (deleted_at);


--
-- Name: idx_app_users_email; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_app_users_email ON public.app_users USING btree (email);


--
-- Name: idx_app_users_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_app_users_project_id ON public.app_users USING btree (project_id);


--
-- Name: idx_app_users_username; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_app_users_username ON public.app_users USING btree (username);


--
-- Name: idx_audit_logs_action; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_audit_logs_action ON public.audit_logs USING btree (action);


--
-- Name: idx_audit_logs_actor_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_audit_logs_actor_id ON public.audit_logs USING btree (actor_id);


--
-- Name: idx_audit_logs_created_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_audit_logs_created_at ON public.audit_logs USING btree (created_at);


--
-- Name: idx_audit_logs_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_audit_logs_deleted_at ON public.audit_logs USING btree (deleted_at);


--
-- Name: idx_audit_logs_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_audit_logs_project_id ON public.audit_logs USING btree (project_id);


--
-- Name: idx_audit_logs_resource_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_audit_logs_resource_id ON public.audit_logs USING btree (resource_id);


--
-- Name: idx_backups_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_backups_deleted_at ON public.backups USING btree (deleted_at);


--
-- Name: idx_backups_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_backups_project_id ON public.backups USING btree (project_id);


--
-- Name: idx_buckets_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_buckets_deleted_at ON public.buckets USING btree (deleted_at);


--
-- Name: idx_channel_members_channel_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_channel_members_channel_id ON public.channel_members USING btree (channel_id);


--
-- Name: idx_channel_members_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_channel_members_deleted_at ON public.channel_members USING btree (deleted_at);


--
-- Name: idx_channel_members_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_channel_members_project_id ON public.channel_members USING btree (project_id);


--
-- Name: idx_channel_members_user_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_channel_members_user_id ON public.channel_members USING btree (user_id);


--
-- Name: idx_channels_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_channels_deleted_at ON public.channels USING btree (deleted_at);


--
-- Name: idx_channels_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_channels_project_id ON public.channels USING btree (project_id);


--
-- Name: idx_channels_type; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_channels_type ON public.channels USING btree (type);


--
-- Name: idx_collections_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_collections_deleted_at ON public.collections USING btree (deleted_at);


--
-- Name: idx_cors_configs_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_cors_configs_deleted_at ON public.cors_configs USING btree (deleted_at);


--
-- Name: idx_cors_configs_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE UNIQUE INDEX idx_cors_configs_project_id ON public.cors_configs USING btree (project_id);


--
-- Name: idx_deployments_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_deployments_deleted_at ON public.deployments USING btree (deleted_at);


--
-- Name: idx_deployments_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_deployments_project_id ON public.deployments USING btree (project_id);


--
-- Name: idx_documents_collection_name; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_documents_collection_name ON public.documents USING btree (collection_name);


--
-- Name: idx_documents_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_documents_deleted_at ON public.documents USING btree (deleted_at);


--
-- Name: idx_documents_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_documents_project_id ON public.documents USING btree (project_id);


--
-- Name: idx_files_bucket_name; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_files_bucket_name ON public.files USING btree (bucket_name);


--
-- Name: idx_files_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_files_deleted_at ON public.files USING btree (deleted_at);


--
-- Name: idx_files_folder_path; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_files_folder_path ON public.files USING btree (folder_path);


--
-- Name: idx_files_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_files_project_id ON public.files USING btree (project_id);


--
-- Name: idx_function_domains_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_function_domains_deleted_at ON public.function_domains USING btree (deleted_at);


--
-- Name: idx_function_domains_domain; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE UNIQUE INDEX idx_function_domains_domain ON public.function_domains USING btree (domain);


--
-- Name: idx_function_domains_function_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_function_domains_function_id ON public.function_domains USING btree (function_id);


--
-- Name: idx_function_domains_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_function_domains_project_id ON public.function_domains USING btree (project_id);


--
-- Name: idx_function_executions_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_function_executions_deleted_at ON public.function_executions USING btree (deleted_at);


--
-- Name: idx_function_executions_execution_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE UNIQUE INDEX idx_function_executions_execution_id ON public.function_executions USING btree (execution_id);


--
-- Name: idx_function_executions_function_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_function_executions_function_id ON public.function_executions USING btree (function_id);


--
-- Name: idx_function_executions_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_function_executions_project_id ON public.function_executions USING btree (project_id);


--
-- Name: idx_functions_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_functions_deleted_at ON public.functions USING btree (deleted_at);


--
-- Name: idx_git_hub_repositories_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_git_hub_repositories_deleted_at ON public.git_hub_repositories USING btree (deleted_at);


--
-- Name: idx_git_hub_repositories_full_name; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_git_hub_repositories_full_name ON public.git_hub_repositories USING btree (full_name);


--
-- Name: idx_git_hub_repositories_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_git_hub_repositories_project_id ON public.git_hub_repositories USING btree (project_id);


--
-- Name: idx_host_key_entries_hostname; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_host_key_entries_hostname ON public.host_key_entries USING btree (hostname);


--
-- Name: idx_host_key_entries_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_host_key_entries_project_id ON public.host_key_entries USING btree (project_id);


--
-- Name: idx_message_reactions_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_message_reactions_deleted_at ON public.message_reactions USING btree (deleted_at);


--
-- Name: idx_message_reactions_message_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_message_reactions_message_id ON public.message_reactions USING btree (message_id);


--
-- Name: idx_message_reactions_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_message_reactions_project_id ON public.message_reactions USING btree (project_id);


--
-- Name: idx_message_reactions_user_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_message_reactions_user_id ON public.message_reactions USING btree (user_id);


--
-- Name: idx_message_reads_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_message_reads_deleted_at ON public.message_reads USING btree (deleted_at);


--
-- Name: idx_message_reads_message_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_message_reads_message_id ON public.message_reads USING btree (message_id);


--
-- Name: idx_message_reads_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_message_reads_project_id ON public.message_reads USING btree (project_id);


--
-- Name: idx_message_reads_user_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_message_reads_user_id ON public.message_reads USING btree (user_id);


--
-- Name: idx_messages_channel_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_messages_channel_id ON public.messages USING btree (channel_id);


--
-- Name: idx_messages_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_messages_deleted_at ON public.messages USING btree (deleted_at);


--
-- Name: idx_messages_parent_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_messages_parent_id ON public.messages USING btree (parent_id);


--
-- Name: idx_messages_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_messages_project_id ON public.messages USING btree (project_id);


--
-- Name: idx_messages_thread_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_messages_thread_id ON public.messages USING btree (thread_id);


--
-- Name: idx_messages_user_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_messages_user_id ON public.messages USING btree (user_id);


--
-- Name: idx_organizations_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_organizations_deleted_at ON public.organizations USING btree (deleted_at);


--
-- Name: idx_project_function_name; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE UNIQUE INDEX idx_project_function_name ON public.functions USING btree (name, project_id);


--
-- Name: idx_projects_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_projects_deleted_at ON public.projects USING btree (deleted_at);


--
-- Name: idx_projects_organization_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_projects_organization_id ON public.projects USING btree (organization_id);


--
-- Name: idx_projects_slug; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_projects_slug ON public.projects USING btree (slug);


--
-- Name: idx_projects_user_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_projects_user_id ON public.projects USING btree (user_id);


--
-- Name: idx_refresh_tokens_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_refresh_tokens_deleted_at ON public.refresh_tokens USING btree (deleted_at);


--
-- Name: idx_refresh_tokens_token; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE UNIQUE INDEX idx_refresh_tokens_token ON public.refresh_tokens USING btree (token);


--
-- Name: idx_refresh_tokens_user_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_refresh_tokens_user_id ON public.refresh_tokens USING btree (user_id);


--
-- Name: idx_ssh_keys_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_ssh_keys_deleted_at ON public.ssh_keys USING btree (deleted_at);


--
-- Name: idx_ssh_keys_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_ssh_keys_project_id ON public.ssh_keys USING btree (project_id);


--
-- Name: idx_users_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_web_servers_deleted_at; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_web_servers_deleted_at ON public.web_servers USING btree (deleted_at);


--
-- Name: idx_web_servers_project_id; Type: INDEX; Schema: public; Owner: cloudbox
--

CREATE INDEX idx_web_servers_project_id ON public.web_servers USING btree (project_id);


--
-- Name: api_keys update_api_keys_updated_at; Type: TRIGGER; Schema: public; Owner: cloudbox
--

CREATE TRIGGER update_api_keys_updated_at BEFORE UPDATE ON public.api_keys FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: backups update_backups_updated_at; Type: TRIGGER; Schema: public; Owner: cloudbox
--

CREATE TRIGGER update_backups_updated_at BEFORE UPDATE ON public.backups FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: channels update_channels_updated_at; Type: TRIGGER; Schema: public; Owner: cloudbox
--

CREATE TRIGGER update_channels_updated_at BEFORE UPDATE ON public.channels FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: cors_configs update_cors_configs_updated_at; Type: TRIGGER; Schema: public; Owner: cloudbox
--

CREATE TRIGGER update_cors_configs_updated_at BEFORE UPDATE ON public.cors_configs FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: deployments update_deployments_updated_at; Type: TRIGGER; Schema: public; Owner: cloudbox
--

CREATE TRIGGER update_deployments_updated_at BEFORE UPDATE ON public.deployments FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: git_hub_repositories update_git_hub_repositories_updated_at; Type: TRIGGER; Schema: public; Owner: cloudbox
--

CREATE TRIGGER update_git_hub_repositories_updated_at BEFORE UPDATE ON public.git_hub_repositories FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: messages update_messages_updated_at; Type: TRIGGER; Schema: public; Owner: cloudbox
--

CREATE TRIGGER update_messages_updated_at BEFORE UPDATE ON public.messages FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: projects update_projects_updated_at; Type: TRIGGER; Schema: public; Owner: cloudbox
--

CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON public.projects FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: ssh_keys update_ssh_keys_updated_at; Type: TRIGGER; Schema: public; Owner: cloudbox
--

CREATE TRIGGER update_ssh_keys_updated_at BEFORE UPDATE ON public.ssh_keys FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: users update_users_updated_at; Type: TRIGGER; Schema: public; Owner: cloudbox
--

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: web_servers update_web_servers_updated_at; Type: TRIGGER; Schema: public; Owner: cloudbox
--

CREATE TRIGGER update_web_servers_updated_at BEFORE UPDATE ON public.web_servers FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: api_keys api_keys_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: backups backups_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.backups
    ADD CONSTRAINT backups_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: channels channels_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.channels
    ADD CONSTRAINT channels_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: cors_configs cors_configs_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.cors_configs
    ADD CONSTRAINT cors_configs_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: deployments deployments_github_repository_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.deployments
    ADD CONSTRAINT deployments_github_repository_id_fkey FOREIGN KEY (github_repository_id) REFERENCES public.git_hub_repositories(id) ON DELETE SET NULL;


--
-- Name: deployments deployments_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.deployments
    ADD CONSTRAINT deployments_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: deployments deployments_web_server_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.deployments
    ADD CONSTRAINT deployments_web_server_id_fkey FOREIGN KEY (web_server_id) REFERENCES public.web_servers(id) ON DELETE SET NULL;


--
-- Name: app_sessions fk_app_sessions_project; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.app_sessions
    ADD CONSTRAINT fk_app_sessions_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: app_users fk_app_users_project; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.app_users
    ADD CONSTRAINT fk_app_users_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: buckets fk_buckets_project; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.buckets
    ADD CONSTRAINT fk_buckets_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: channels fk_channels_project; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.channels
    ADD CONSTRAINT fk_channels_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: collections fk_collections_project; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.collections
    ADD CONSTRAINT fk_collections_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: deployments fk_deployments_git_hub_repository; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.deployments
    ADD CONSTRAINT fk_deployments_git_hub_repository FOREIGN KEY (git_hub_repository_id) REFERENCES public.git_hub_repositories(id);


--
-- Name: deployments fk_deployments_web_server; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.deployments
    ADD CONSTRAINT fk_deployments_web_server FOREIGN KEY (web_server_id) REFERENCES public.web_servers(id);


--
-- Name: function_domains fk_function_domains_function; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.function_domains
    ADD CONSTRAINT fk_function_domains_function FOREIGN KEY (function_id) REFERENCES public.functions(id);


--
-- Name: function_domains fk_function_domains_project; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.function_domains
    ADD CONSTRAINT fk_function_domains_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: function_executions fk_function_executions_function; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.function_executions
    ADD CONSTRAINT fk_function_executions_function FOREIGN KEY (function_id) REFERENCES public.functions(id);


--
-- Name: function_executions fk_function_executions_project; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.function_executions
    ADD CONSTRAINT fk_function_executions_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: functions fk_functions_project; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.functions
    ADD CONSTRAINT fk_functions_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: git_hub_repositories fk_git_hub_repositories_project; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.git_hub_repositories
    ADD CONSTRAINT fk_git_hub_repositories_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: git_hub_repositories fk_git_hub_repositories_ssh_key; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.git_hub_repositories
    ADD CONSTRAINT fk_git_hub_repositories_ssh_key FOREIGN KEY (ssh_key_id) REFERENCES public.ssh_keys(id);


--
-- Name: organizations fk_organizations_user; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT fk_organizations_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: api_keys fk_projects_api_keys; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT fk_projects_api_keys FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: backups fk_projects_backups; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.backups
    ADD CONSTRAINT fk_projects_backups FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: cors_configs fk_projects_cors_config; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.cors_configs
    ADD CONSTRAINT fk_projects_cors_config FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: deployments fk_projects_deployments; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.deployments
    ADD CONSTRAINT fk_projects_deployments FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: projects fk_projects_organization; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_organization FOREIGN KEY (organization_id) REFERENCES public.organizations(id);


--
-- Name: ssh_keys fk_ssh_keys_project; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.ssh_keys
    ADD CONSTRAINT fk_ssh_keys_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: projects fk_users_projects; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_users_projects FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: refresh_tokens fk_users_refresh_tokens; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT fk_users_refresh_tokens FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: web_servers fk_web_servers_project; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.web_servers
    ADD CONSTRAINT fk_web_servers_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: web_servers fk_web_servers_ssh_key; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.web_servers
    ADD CONSTRAINT fk_web_servers_ssh_key FOREIGN KEY (ssh_key_id) REFERENCES public.ssh_keys(id);


--
-- Name: git_hub_repositories git_hub_repositories_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.git_hub_repositories
    ADD CONSTRAINT git_hub_repositories_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: messages messages_channel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_channel_id_fkey FOREIGN KEY (channel_id) REFERENCES public.channels(id) ON DELETE CASCADE;


--
-- Name: messages messages_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: projects projects_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: ssh_keys ssh_keys_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.ssh_keys
    ADD CONSTRAINT ssh_keys_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: web_servers web_servers_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.web_servers
    ADD CONSTRAINT web_servers_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: web_servers web_servers_ssh_key_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: cloudbox
--

ALTER TABLE ONLY public.web_servers
    ADD CONSTRAINT web_servers_ssh_key_id_fkey FOREIGN KEY (ssh_key_id) REFERENCES public.ssh_keys(id) ON DELETE SET NULL;


--
-- PostgreSQL database dump complete
--

