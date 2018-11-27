FROM mongo:3.6-jessie

RUN set -x \
  && apt-get update \
  && apt-get install -y --no-install-recommends \
    ca-certificates \
    netcat \
  && rm -rf /var/lib/apt/lists/* /usr/share/doc /usr/share/man /tmp/*

COPY osm /usr/local/bin/osm
COPY mongo-tools.sh /usr/local/bin/mongo-tools.sh

ENTRYPOINT ["mongo-tools.sh"]
