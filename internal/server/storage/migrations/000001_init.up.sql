CREATE TABLE users (
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    password   TEXT NOT NULL,
    is_disable BOOLEAN NOT NULL DEFAULT FALSE,
    update_id  INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE data (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    name  TEXT NOT NULL,
    value  TEXT NOT NULL,
    created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_user_data_name ON data(user_id, name);

insert into users (name, password) values ('user', '5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8');
