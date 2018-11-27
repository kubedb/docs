#!/bin/bash

mkdir -p "$PGDATA"
rm -rf "$PGDATA"/*
chmod 0700 "$PGDATA"

# set wal-g ENV
CRED_PATH="/srv/wal-g/restore/secrets"
export WALE_S3_PREFIX=$(echo "$RESTORE_S3_PREFIX")
export AWS_ACCESS_KEY_ID=$(cat "$CRED_PATH/AWS_ACCESS_KEY_ID")
export AWS_SECRET_ACCESS_KEY=$(cat "$CRED_PATH/AWS_SECRET_ACCESS_KEY")

PITR=${PITR:-false}
TARGET_INCLUSIVE=${TARGET_INCLUSIVE:-true}
TARGET_TIME=${TARGET_TIME:-}
TARGET_TIMELINE=${TARGET_TIMELINE:-}
TARGET_XID=${TARGET_XID:-}

until wal-g backup-list &>/dev/null; do
  echo "waiting for archived backup..."
  sleep 5
done

echo "Fetching archived backup..."
# fetch backup
wal-g backup-fetch "$PGDATA" "$BACKUP_NAME" >/dev/null

# create missing folders
mkdir -p "$PGDATA"/{pg_tblspc,pg_twophase,pg_stat,pg_commit_ts}/
mkdir -p "$PGDATA"/pg_logical/{snapshots,mappings}/

# setup recovery.conf
cp /scripts/replica/recovery.conf /tmp

# ref: https://www.postgresql.org/docs/10/static/recovery-target-settings.html
if [ "$PITR" = true ]; then
  echo "recovery_target_inclusive = '$TARGET_INCLUSIVE'" >>/tmp/recovery.conf
  echo "recovery_target_action = 'promote'" >>/tmp/recovery.conf

  if [ ! -z "$TARGET_TIME" ]; then
    echo "recovery_target_time = '$TARGET_TIME'" >>/tmp/recovery.conf
  fi
  if [ ! -z "$TARGET_TIMELINE" ]; then
    echo "recovery_target_timeline = '$TARGET_TIMELINE'" >>/tmp/recovery.conf
  fi
  if [ ! -z "$TARGET_XID" ]; then
    echo "recovery_target_xid = '$TARGET_XID'" >>/tmp/recovery.conf
  fi
fi

echo "restore_command = 'wal-g wal-fetch %f %p'" >>/tmp/recovery.conf
mv /tmp/recovery.conf "$PGDATA/recovery.conf"

# setup postgresql.conf
cp /scripts/primary/postgresql.conf /tmp
echo "wal_level = replica" >>/tmp/postgresql.conf
echo "max_wal_senders = 99" >>/tmp/postgresql.conf
echo "wal_keep_segments = 32" >>/tmp/postgresql.conf
if [ "$STREAMING" == "synchronous" ]; then
  # setup synchronous streaming replication
  echo "synchronous_commit = remote_write" >>/tmp/postgresql.conf
  echo "synchronous_standby_names = '*'" >>/tmp/postgresql.conf
fi
mv /tmp/postgresql.conf "$PGDATA/postgresql.conf"

if [ "$ARCHIVE" == "wal-g" ]; then
  # setup postgresql.conf
  echo "archive_command = 'wal-g wal-push %p'" >>"$PGDATA/postgresql.conf"
  echo "archive_timeout = 60" >>"$PGDATA/postgresql.conf"
  echo "archive_mode = always" >>"$PGDATA/postgresql.conf"
fi

rm "$PGDATA/recovery.done" &>/dev/null

# start server for recovery process
pg_ctl -D "$PGDATA" -W start >/dev/null

# this file will trigger recovery
touch '/tmp/pg-failover-trigger'

# This will hold until recovery completed
while [ ! -e "$PGDATA/recovery.done" ]; do
  echo "replaying wal files..."
  sleep 5
done

# create PID if misssing
postmaster -D "$PGDATA" &>/dev/null

pg_ctl -D "$PGDATA" -w stop >/dev/null
