CREATE DATABASE smart_home_assistant;
CREATE TABLE devices (
    device_id VARCHAR PRIMARY KEY,
    type VARCHAR NOT NULL,
    state VARCHAR NOT NULL
);

CREATE USER your_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE smart_home_assistant TO your_user;
