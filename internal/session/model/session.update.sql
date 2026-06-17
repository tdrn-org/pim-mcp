UPDATE
    session
SET
    api_key_shown = $1,
    credentials = $2,
    last_update = $3
WHERE
    id = $4