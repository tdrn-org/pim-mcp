--
-- Devices
--
CREATE TABLE session(
    id TEXT NOT NULL,
    api_key TEXT NOT NULL,
    secrets TEXT NOT NULL,
    last_access INTEGER NOT NULL,
    PRIMARY KEY(id)
);
--
-- EOF
--
