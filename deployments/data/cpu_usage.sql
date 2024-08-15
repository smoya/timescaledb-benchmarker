SELECT 'CREATE DATABASE homework'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'homework')\gexec
\c homework
CREATE EXTENSION IF NOT EXISTS timescaledb;
CREATE TABLE cpu_usage(
  ts    TIMESTAMPTZ,
  host  TEXT,
  usage DOUBLE PRECISION
);
SELECT create_hypertable('cpu_usage', 'ts');
COPY cpu_usage FROM '/docker-entrypoint-initdb.d/cpu_usage.csv' CSV HEADER;