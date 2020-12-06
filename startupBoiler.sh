#!/usr/bin/env bash

# -n to ignore running tests after generation
# -p to preserve old database

stop_containers () {
    echo "[MOEBOT] Bringing down database container..."
    # Cleanup our previous state
    if docker-compose -f "docker-compose.yml" -f "docker-compose.dev.yml" down; then
        echo "[MOEBOT] Successfully brought down the database container."
    else
        echo "[MOEBOT] ERROR!"
        echo "[MOEBOT] Failed to destroy database container. See above errors"
        echo "[MOEBOT] ERROR!"
        exit
    fi
}

print_logs () {
    docker logs db
}

stop_containers

if ! [[ $1 =~ "p" ]]; then
    echo "[MOEBOT] Removing old volume so the database will re-init"
    docker volume remove moebot-data
    docker volume create moebot-data
else
    echo "[MOEBOT] Skipping removal of old database volume"
fi

echo "[MOEBOT] Starting up database docker container..."
# Startup our docker-compose, but including the dev version so we expose the db port
if docker-compose -f "docker-compose.yml" -f "docker-compose.dev.yml" up --build -d db; then
    echo "[MOEBOT] Successfully created database container"
else
    echo "[MOEBOT] ERROR!"
    echo "[MOEBOT] Failed to create database container. See above errors"
    echo "[MOEBOT] ERROR!"
    exit
fi

echo "[MOEBOT] Waiting for database to init..."
# Sleep an initial waiting period
sleep 10
# Then go into our loop trying to pint the server
RETRIES=10
until docker logs db 2>&1 | grep 'listening on IPv4 address' || [[ ${RETRIES} -eq 0 ]]; do
  echo "Waiting for postgres server, $((RETRIES--)) remaining attempts..."
  sleep 5
done

echo "[MOEBOT] Starting SqlBoiler model generation..."
# Then build our models
if sqlboiler psql --wipe --no-hooks -o ./moebot_bot/util/db/models/; then
    echo "[MOEBOT] Successfully built SqlBoiler models."
else
    echo "[MOEBOT] ERROR!"
    echo "[MOEBOT] Failed to build SqlBoiler models. See above errors"
    echo "[MOEBOT] The startup script will continue so that docker containers can be cleaned up."
    echo "[MOEBOT] ERROR!"
    print_logs
    stop_containers
    exit
fi

if ! [[ $1 =~ "n" ]]; then
    echo "[MOEBOT] Running SqlBoiler tests"
    if ! go test ./moebot_bot/util/db/models; then
        echo "[MOEBOT] Model testing failed!"
    fi
fi

stop_containers
