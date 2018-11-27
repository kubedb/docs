#!/usr/bin/env bash

set -e

echo "Running as Replica"

mkdir -p "$PGDATA"
rm -rf "$PGDATA"/*
chmod 0700 "$PGDATA"

# set password ENV
export PGPASSWORD=${POSTGRES_PASSWORD:-postgres}

export ARCHIVE=${ARCHIVE:-}

# Waiting for running Postgres
while true; do
  pg_isready --host="$PRIMARY_HOST" --timeout=2 &>/dev/null && break
  echo "Attempting pg_isready on primary"
  sleep 2
done
while true; do
  psql -h "$PRIMARY_HOST" --no-password --username=postgres --command="select now();" &>/dev/null && break
  echo "Attempting query on primary"
  sleep 2
done

# get basebackup
pg_basebackup -X fetch --no-password --pgdata "$PGDATA" --username=postgres --host="$PRIMARY_HOST"

# setup recovery.conf
cp /scripts/replica/recovery.conf /tmp
echo "recovery_target_timeline = 'latest'" >>/tmp/recovery.conf
echo "archive_cleanup_command = 'pg_archivecleanup $PGWAL %r'" >>/tmp/recovery.conf
# primary_conninfo is used for streaming replication
echo "primary_conninfo = 'application_name=$HOSTNAME host=$PRIMARY_HOST'" >>/tmp/recovery.conf
mv /tmp/recovery.conf "$PGDATA/recovery.conf"

# setup postgresql.conf
cp /scripts/primary/postgresql.conf /tmp
echo "wal_level = replica" >>/tmp/postgresql.conf
echo "max_wal_senders = 99" >>/tmp/postgresql.conf
echo "wal_keep_segments = 32" >>/tmp/postgresql.conf
if [ "$STANDBY" == "hot" ]; then
  echo "hot_standby = on" >>/tmp/postgresql.conf
fi
if [ "$STREAMING" == "synchronous" ]; then
   # setup synchronous streaming replication
   echo "synchronous_commit = remote_write" >>/tmp/postgresql.conf
   echo "synchronous_standby_names = '*'" >>/tmp/postgresql.conf
fi
mv /tmp/postgresql.conf "$PGDATA/postgresql.conf"

# push base-backup
if [ "$ARCHIVE" == "wal-g" ]; then
  # set walg ENV
  CRED_PATH="/srv/wal-g/archive/secrets"
  export WALE_S3_PREFIX=$(echo "$ARCHIVE_S3_PREFIX")
  export AWS_ACCESS_KEY_ID=$(cat "$CRED_PATH/AWS_ACCESS_KEY_ID")
  export AWS_SECRET_ACCESS_KEY=$(cat "$CRED_PATH/AWS_SECRET_ACCESS_KEY")

  # setup postgresql.conf
  echo "archive_command = 'wal-g wal-push %p'" >>"$PGDATA/postgresql.conf"
  echo "archive_timeout = 60" >>"$PGDATA/postgresql.conf"
  echo "archive_mode = always" >>"$PGDATA/postgresql.conf"
fi

exec postgres
