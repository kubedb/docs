FROM quay.io/coreos/etcd:v3.2.13

RUN set -x \
  && apk add --update --no-cache ca-certificates

COPY osm /usr/local/bin/osm
COPY etcd-tools.sh /usr/local/bin/etcd-tools.sh

ENTRYPOINT ["etcd-tools.sh"]
