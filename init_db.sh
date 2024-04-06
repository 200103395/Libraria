#!/bin/bash

echo "Start initialization db bash script"
# Start the postgres service with listen_addresses set to 0.0.0.0
postgres -c "listen_addresses=0.0.0.0" &

# Wait for database initialization (adjust for your setup)
sleep 5

# Execute psql to restore the database from libraria.sql
export PGPASSWORD=$POSTGRES_PASSWORD
psql -h $HOST -U $POSTGRES_USER -d $POSTGRES_DB -f /docker-entrypoint-initdb.d/libraria.sql

echo "Database initialization complete!"
