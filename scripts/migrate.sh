#!/bin/sh

# This script run all the migrations. It is used in production environment.
migrate -path "./scripts/migrations" -database "postgres://$DB_USERNAME:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSL_MODE" -verbose up