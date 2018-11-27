FROM mysql:8.0.3

RUN set -x \
  && apt-get update \
  && apt-get install -y --no-install-recommends \
    ca-certificates \
    netcat \
  && rm -rf /var/lib/apt/lists/* /usr/share/doc /usr/share/man /tmp/*

COPY osm /usr/local/bin/osm
COPY mysql-tools.sh /usr/local/bin/mysql-tools.sh

ENTRYPOINT ["mysql-tools.sh"]
