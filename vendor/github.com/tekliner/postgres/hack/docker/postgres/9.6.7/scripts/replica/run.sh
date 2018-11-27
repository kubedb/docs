#!/usr/bin/env bash

set -e

echo "replica run.sh:5 Running as Replica"

mkdir -p "$PGDATA"
rm -rf "$PGDATA"/*
chmod 0700 "$PGDATA"

# set password ENV
export PGPASSWORD=${POSTGRES_PASSWORD:-postgres}

export ARCHIVE=${ARCHIVE:-}

echo "replica run.sh:16 get basebackup"
pg_basebackup -X fetch --no-password --pgdata "$PGDATA" --username=postgres --host="$PRIMARY_HOST"

echo "replica run.sh:19 setup recovery.conf"
cp /scripts/replica/recovery.conf /tmp
echo "recovery_target_timeline = 'latest'" >>/tmp/recovery.conf
echo "archive_cleanup_command = 'pg_archivecleanup $PGWAL %r'" >>/tmp/recovery.conf
# primary_conninfo is used for streaming replication
echo "primary_conninfo = 'application_name=$HOSTNAME host=$PRIMARY_HOST'" >>/tmp/recovery.conf
mv /tmp/recovery.conf "$PGDATA/recovery.conf"

echo "replica run.sh:27 setup postgresql.conf"
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

echo "replica run.sh:42 push base-backup"
if [ "$ARCHIVE" == "wal-g" ]; then
  # set walg ENV
  CRED_PATH="/srv/wal-g/archive/secrets"
  export WALE_S3_PREFIX=$(echo "$ARCHIVE_S3_PREFIX")
  export AWS_ACCESS_KEY_ID=$(cat "$CRED_PATH/AWS_ACCESS_KEY_ID")
  export AWS_SECRET_ACCESS_KEY=$(cat "$CRED_PATH/AWS_SECRET_ACCESS_KEY")

  echo "replica run.sh:50 setup postgresql.conf"
  echo "archive_command = 'wal-g wal-push %p'" >>"$PGDATA/postgresql.conf"
  echo "archive_timeout = 60" >>"$PGDATA/postgresql.conf"
  echo "archive_mode = always" >>"$PGDATA/postgresql.conf"
fi

echo "replica run.sh:56 exec postgres"
exec postgres
