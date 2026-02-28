#!/bin/bash
# Creates functionfly_blog for the NestJS blog API. Runs in Postgres container init.
set -e
psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d postgres -c "CREATE DATABASE functionfly_blog;"
