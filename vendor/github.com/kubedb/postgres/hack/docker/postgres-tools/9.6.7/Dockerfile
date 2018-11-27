FROM postgres:9.6.7-alpine

RUN set -x \
  && apk add --update --no-cache ca-certificates

COPY osm /usr/local/bin/osm
COPY postgres-tools.sh /usr/local/bin/postgres-tools.sh

ENTRYPOINT ["postgres-tools.sh"]
