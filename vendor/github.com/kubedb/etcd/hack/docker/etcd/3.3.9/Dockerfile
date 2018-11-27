FROM quay.io/coreos/etcd:v3.3.9

RUN set -x \
  && apk add --update --no-cache ca-certificates


COPY etcd-operator /usr/bin/

ENTRYPOINT ["etcd-operator"]
CMD ["etcd-helper"]