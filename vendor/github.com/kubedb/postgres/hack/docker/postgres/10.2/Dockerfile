FROM postgres:10.2-alpine

RUN set -x \
  && apk add --update --no-cache ca-certificates

ENV PV /var/pv
ENV PGDATA $PV/data
ENV PGWAL $PGDATA/pg_wal
ENV INITDB /var/initdb
ENV WALG_D /etc/wal-g.d/env

COPY pg-operator /usr/bin/
COPY wal-g /usr/bin/

COPY scripts /scripts

VOLUME ["$PV"]

ENV STANDBY warm
ENV RESTORE false
ENV BACKUP_NAME LATEST
ENV PITR false
ENV ARCHIVE_S3_PREFIX ""
ENV RESTORE_S3_PREFIX ""

ENTRYPOINT ["pg-operator"]
CMD ["leader_election"]

EXPOSE 5432
