SELECT
    a.id,
    a.api_key_shown,
    a.credentials,
    a.last_access
FROM
    session a
WHERE
    a.api_key = $1
