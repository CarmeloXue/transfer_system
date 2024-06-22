-- create_databases.sql

DO
$do$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'yourdb') THEN
      PERFORM dblink_exec('dbname=postgres', 'CREATE DATABASE yourdb');
   END IF;
END
$do$;
