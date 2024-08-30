CREATE KEYSPACE IF NOT EXISTS pipe WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1} AND durable_writes = true;
USE pipe;

CREATE TABLE IF NOT EXISTS users_by_id (
    user_id BIGINT PRIMARY KEY,
    private_id TEXT,
    pubkey TEXT,
    created_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users_by_private_id (
    private_id TEXT PRIMARY KEY,
    user_id BIGINT,
    pubkey TEXT,
    created_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS messages (
    message_id UUID,
    from_user BIGINT,
    to_user BIGINT,
    text TEXT,
    date BIGINT,
    PRIMARY KEY (to_user, date, message_id)
) WITH CLUSTERING ORDER BY (date DESC);
