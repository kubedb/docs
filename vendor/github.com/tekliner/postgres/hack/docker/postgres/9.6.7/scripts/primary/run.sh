#!/usr/bin/env bash

set -e

echo "master: run.sh:5 Running as Primary"

# set password ENV
export PGPASSWORD=${POSTGRES_PASSWORD:-postgres}

export ARCHIVE=${ARCHIVE:-}

if [ ! -e "$PGDATA/PG_VERSION" ]; then
  if [ "$RESTORE" = true ]; then
    echo "master: run.sh:14 Restoring Postgres from base_backup using wal-g"
    /scripts/primary/restore.sh
  else
    /scripts/primary/start.sh
  fi
fi

echo "master: run.sh:21 push base-backup"
if [ "$ARCHIVE" == "wal-g" ]; then
  # set walg ENV
  CRED_PATH="/srv/wal-g/archive/secrets"
  export WALE_S3_PREFIX=$(echo "$ARCHIVE_S3_PREFIX")
  export AWS_ACCESS_KEY_ID=$(cat "$CRED_PATH/AWS_ACCESS_KEY_ID")
  export AWS_SECRET_ACCESS_KEY=$(cat "$CRED_PATH/AWS_SECRET_ACCESS_KEY")

  echo "master: run.sh:29 pg_ctl -w start"
  pg_ctl -D "$PGDATA" -w start

  echo "master: run.sh:32 Wait for connection"
  while ! psql -U postgres -c "select 1;" > /dev/null; do
    echo "run.sh:34 Connection failed"
    sleep 5
  done

  echo "master: run.sh:38 start wal-g"
  PGUSER="postgres" wal-g backup-push "$PGDATA" >/dev/null

  echo "master: run.sh:41 Set $POSTGRES_USER password for remote connections if DB restored from alien backup"
  psql -U postgres -c "ALTER USER $POSTGRES_USER WITH PASSWORD '$POSTGRES_PASSWORD';"

  echo "master: run.sh:44 pg_ctl stop"
  pg_ctl -D "$PGDATA" -m fast -w stop
fi

echo "master: run.sh:48 exec postgres"
exec postgres
