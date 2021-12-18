#!/bin/sh

echo "run db migration"
migrate -path /app/migrations -database "$POSTGRES_URL" -verbose up
