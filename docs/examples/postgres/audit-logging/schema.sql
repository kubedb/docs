-- Audit plumbing -----------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS pgaudit;

-- Never logged into; GRANTs to this role mark what pgAudit records reads of.
CREATE ROLE auditor NOLOGIN;

-- Example core banking schema ---------------------------------------------
CREATE SCHEMA core;

CREATE TABLE core.customers (
    id            bigserial PRIMARY KEY,
    full_name     text NOT NULL,
    national_id   text NOT NULL,
    date_of_birth date NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE core.accounts (
    id          bigserial PRIMARY KEY,
    account_no  text UNIQUE NOT NULL,
    customer_id bigint NOT NULL REFERENCES core.customers(id),
    balance     numeric(18,2) NOT NULL DEFAULT 0,
    status      text NOT NULL DEFAULT 'active'
);

CREATE TABLE core.card_details (
    id         bigserial PRIMARY KEY,
    account_id bigint NOT NULL REFERENCES core.accounts(id),
    pan        text NOT NULL,
    expiry     date NOT NULL
);

CREATE TABLE core.transactions (
    id         bigserial PRIMARY KEY,
    account_id bigint NOT NULL REFERENCES core.accounts(id),
    amount     numeric(18,2) NOT NULL,
    kind       text NOT NULL,
    booked_at  timestamptz NOT NULL DEFAULT now()
);

-- Object auditing: nominate the sensitive data ----------------------------
GRANT SELECT (national_id, date_of_birth) ON core.customers TO auditor;
GRANT SELECT ON core.card_details TO auditor;

-- Named actors, so audit records attribute actions -------------------------
CREATE ROLE teller LOGIN PASSWORD 'CHANGE-ME';
CREATE ROLE payments_svc LOGIN PASSWORD 'CHANGE-ME';

GRANT USAGE ON SCHEMA core TO teller, payments_svc;
GRANT SELECT ON core.customers, core.accounts, core.card_details TO teller;
GRANT SELECT, INSERT ON core.transactions TO payments_svc;
GRANT SELECT, UPDATE ON core.accounts TO payments_svc;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA core TO payments_svc;
