#!/bin/bash

mkdir -p "$PGDATA"
rm -rf "$PGDATA"/*
chmod 0700 "$PGDATA"

export POSTGRES_INITDB_ARGS=${POSTGRES_INITDB_ARGS:-}
export POSTGRES_INITDB_WALDIR=${POSTGRES_INITDB_WALDIR:-}

# Create the transaction log directory before initdb is run
if [ "$POSTGRES_INITDB_WALDIR" ]; then
  mkdir -p "$POSTGRES_INITDB_WALDIR"
  chown -R postgres "$POSTGRES_INITDB_WALDIR"
  chmod 700 "$POSTGRES_INITDB_WALDIR"

  export POSTGRES_INITDB_ARGS="$POSTGRES_INITDB_ARGS --waldir $POSTGRES_INITDB_WALDIR"
fi

initdb $POSTGRES_INITDB_ARGS --pgdata="$PGDATA" >/dev/null

# setup postgresql.conf
cp /scripts/primary/postgresql.conf /tmp
echo "wal_level = replica" >>/tmp/postgresql.conf
echo "max_wal_senders = 99" >>/tmp/postgresql.conf
echo "wal_keep_segments = 32" >>/tmp/postgresql.conf
mv /tmp/postgresql.conf "$PGDATA/postgresql.conf"

# setup pg_hba.conf
{ echo; echo 'local all         all                         trust'; }   >>"$PGDATA/pg_hba.conf"
{       echo 'host  all         all         127.0.0.1/32    trust'; }   >>"$PGDATA/pg_hba.conf"
{       echo 'host  all         all         0.0.0.0/0       md5'; }     >>"$PGDATA/pg_hba.conf"
{       echo 'host  replication postgres    0.0.0.0/0       md5'; }     >>"$PGDATA/pg_hba.conf"

# start postgres
pg_ctl -D "$PGDATA" -w start >/dev/null

export POSTGRES_USER=${POSTGRES_USER:-postgres}
export POSTGRES_DB=${POSTGRES_DB:-$POSTGRES_USER}
export POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-postgres}

psql=(psql -v ON_ERROR_STOP=1)

# create database with specified name
if [ "$POSTGRES_DB" != "postgres" ]; then
  "${psql[@]}" --username postgres <<-EOSQL
CREATE DATABASE "$POSTGRES_DB" ;
EOSQL
  echo
fi

if [ "$POSTGRES_USER" = "postgres" ]; then
  op="ALTER"
else
  op="CREATE"
fi

# alter postgres superuser
"${psql[@]}" --username postgres <<-EOSQL
    $op USER "$POSTGRES_USER" WITH SUPERUSER PASSWORD '$POSTGRES_PASSWORD';
EOSQL
echo

psql+=(--username "$POSTGRES_USER" --dbname "$POSTGRES_DB")
echo

# initialize database
for f in "$INITDB"/*; do
  case "$f" in
    *.sh)     echo "$0: running $f"; . "$f" ;;
    *.sql)    echo "$0: running $f"; "${psql[@]}" -f "$f"; echo ;;
    *.sql.gz) echo "$0: running $f"; gunzip -c "$f" | "${psql[@]}"; echo ;;
    *)        echo "$0: ignoring $f" ;;
  esac
  echo
done

# stop server
pg_ctl -D "$PGDATA" -m fast -w stop >/dev/null

if [ "$STREAMING" == "synchronous" ]; then
   # setup synchronous streaming replication
   echo "synchronous_commit = remote_write" >>"$PGDATA/postgresql.conf"
   echo "synchronous_standby_names = '*'" >>"$PGDATA/postgresql.conf"
fi


if [ "$ARCHIVE" == "wal-g" ]; then
  # setup postgresql.conf
  echo "archive_command = 'wal-g wal-push %p'" >>"$PGDATA/postgresql.conf"
  echo "archive_timeout = 60" >>"$PGDATA/postgresql.conf"
  echo "archive_mode = always" >>"$PGDATA/postgresql.conf"
fi
