-- +goose Up
CREATE TABLE root_certificates (
                                    id                          uuid PRIMARY KEY,
                                    name                        text NOT NULL UNIQUE CHECK (name <> ''),
                                    key                         bytea NOT NULL,
                                    cert                        bytea NOT NULL,
                                    crl                         bytea NOT NULL DEFAULT '',
                                    not_before                  timestamptz NOT NULL,
                                    not_after                   timestamptz NOT NULL,
                                    client_default_valid_days   int NOT NULL
);

CREATE TABLE dh (
                                    id          uuid PRIMARY KEY,
                                    name        text NOT NULL UNIQUE CHECK (name <> ''),
                                    der_bytes   bytea NOT NULL
);

CREATE TABLE server_certificates (
                                     id          uuid PRIMARY KEY,
                                     name        text NOT NULL CHECK (name <> ''),
                                     key         bytea NOT NULL,
                                     cert        bytea NOT NULL,
                                     not_before  timestamptz NOT NULL,
                                     not_after   timestamptz NOT NULL,
                                     root_id       uuid NOT NULL,
                                     is_revoked  boolean NOT NULL DEFAULT false,

                                     CONSTRAINT server_certificate_root_fk
                                         FOREIGN KEY (root_id)
                                             REFERENCES root_certificates(id)
                                             ON DELETE RESTRICT
);

CREATE INDEX server_certificates_root_id_idx
    ON server_certificates (root_id);

CREATE UNIQUE INDEX server_certificates_active_unique
    ON server_certificates (root_id, name)
    WHERE is_revoked = false;

CREATE TABLE client_certificates (
                                     id          uuid PRIMARY KEY,
                                     name        text NOT NULL CHECK (name <> ''),
                                     key         bytea NOT NULL,
                                     cert        bytea NOT NULL,
                                     not_before  timestamptz NOT NULL,
                                     not_after   timestamptz NOT NULL,
                                     root_id       uuid NOT NULL,
                                     is_revoked  boolean NOT NULL DEFAULT false,

                                     CONSTRAINT client_certificate_root_fk
                                         FOREIGN KEY (root_id)
                                             REFERENCES root_certificates(id)
                                             ON DELETE RESTRICT
);

CREATE INDEX client_certificates_root_id_idx
    ON client_certificates (root_id);

CREATE TABLE server_configs (
                                id                      uuid PRIMARY KEY,
                                name                    text NOT NULL UNIQUE CHECK (name <> ''),
                                root_certificate_id     uuid NOT NULL,
                                server_certificate_id   uuid NOT NULL,
                                dh_id                   uuid NOT NULL,
                                server_config_template  text NOT NULL DEFAULT '',
                                client_config_template  text NOT NULL DEFAULT '',
                                groups_allowed          text NOT NULL DEFAULT '',

                                CONSTRAINT server_config_root_certificate_fk
                                    FOREIGN KEY (root_certificate_id)
                                        REFERENCES root_certificates(id)
                                        ON DELETE RESTRICT,

                                CONSTRAINT server_config_server_certificate_fk
                                    FOREIGN KEY (server_certificate_id)
                                        REFERENCES server_certificates(id)
                                        ON DELETE RESTRICT,

                                CONSTRAINT server_config_dh_fk
                                    FOREIGN KEY (dh_id)
                                        REFERENCES dh(id)
                                        ON DELETE RESTRICT
);

CREATE TABLE nodes (
                       name      text PRIMARY KEY,
                       last_seen timestamptz NOT NULL
);

CREATE TABLE server_states (
                               server_id        uuid          NOT NULL,
                               node_name        text          NOT NULL CHECK (node_name <> ''),
                               last_seen        timestamptz   NOT NULL DEFAULT 'epoch'::timestamptz,
                               last_pid         int           NOT NULL DEFAULT 0,
                               reload_scheduled bool          NOT NULL DEFAULT false,

                               PRIMARY KEY (node_name, server_id),

                               CONSTRAINT server_state_server_config_fk
                                   FOREIGN KEY (server_id)
                                       REFERENCES server_configs(id)
                                       ON DELETE CASCADE,

                               CONSTRAINT server_state_node_fk
                                   FOREIGN KEY (node_name)
                                       REFERENCES nodes(name)
                                       ON DELETE CASCADE
);

-- +goose Down
DROP TABLE server_states;
DROP TABLE nodes;
DROP TABLE server_configs;
DROP TABLE client_certificates;
DROP TABLE server_certificates;
DROP TABLE root_certificates;
