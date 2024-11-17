#!/bin/bash

# Database connection details
DB_NAME="postgres"
DB_USER="postgres"
DB_HOST="localhost"
DB_PORT="5432"

# SQL commands to create tables
SQL_COMMANDS="
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(20) NOT NULL,
    password VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS projects (
    id            SERIAL PRIMARY KEY,
    directories   JSONB,
    project_name  VARCHAR(100) NOT NULL,
    user_id       INTEGER NOT NULL,
    created_at    BIGINT NOT NULL,
    mtime         BIGINT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS files (
    id            SERIAL PRIMARY KEY,
    name          VARCHAR(20),
    path          VARCHAR(150),
    data          BYTEA NOT NULL,
    user_id       INTEGER NOT NULL,
    project_id    INTEGER NOT NULL,
    mtime         BIGINT NOT NULL,
    created_at    BIGINT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS shared (
    id            SERIAL PRIMARY KEY,
    hash          VARCHAR(100) NOT NULL,
    user_id       INTEGER NOT NULL,
    project_id    INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
"

# Execute SQL commands using psql
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "$SQL_COMMANDS"