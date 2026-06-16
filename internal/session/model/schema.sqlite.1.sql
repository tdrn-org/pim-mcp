--
-- Devices
--
CREATE TABLE session(
    id TEXT NOT NULL,
    api_key TEXT NOT NULL,
    api_key_shown INTEGER NOT NULL,
    credentials TEXT NOT NULL,
    last_access INTEGER NOT NULL,
    PRIMARY KEY(id),
    UNIQUE(api_key)
);
--
-- EOF
--
