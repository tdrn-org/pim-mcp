SELECT
    a.api_key,
    a.api_key_shown,
    a.credentials,
    a.last_access
FROM
    session a
WHERE
    a.id = $1
