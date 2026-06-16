SELECT
    a.api_key,
    a.secrets,
    a.last_access
FROM
    session a
WHERE
    a.id = $1
