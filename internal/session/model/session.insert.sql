INSERT INTO
    session(
        id,
        api_key,
        api_key_shown,
        secrets,
        last_access
    )
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5
)