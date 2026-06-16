UPDATE
    session
SET
    secrets = $1,
    last_access = $2
WHERE
    id = $3