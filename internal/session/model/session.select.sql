SELECT
    a.secrets
FROM
    session a
WHERE
    a.id = $1
