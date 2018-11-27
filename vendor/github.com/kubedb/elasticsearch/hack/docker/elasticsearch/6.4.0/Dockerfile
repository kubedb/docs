FROM quay.io/pires/docker-elasticsearch:6.4.0

RUN set -x \
  && apk add --update --no-cache runit curl

ENV NODE_NAME="" \
    ES_TMPDIR="/tmp"

# Install mapper-attachments (https://www.elastic.co/guide/en/elasticsearch/plugins/current/ingest-attachment.html)
RUN ./bin/elasticsearch-plugin install --batch ingest-attachment

# Install search-guard
RUN ./bin/elasticsearch-plugin install --batch -b com.floragunn:search-guard-6:6.4.0-23.1

RUN chmod +x -R plugins/search-guard-6/tools/*.sh

# Set environment variables defaults
ENV ES_JAVA_OPTS="-Xms512m -Xmx512m" \
    CLUSTER_NAME="elasticsearch" \
    NODE_MASTER=true \
    NODE_DATA=true \
    NODE_INGEST=true \
    HTTP_ENABLE=true \
    HTTP_CORS_ENABLE=true \
    HTTP_CORS_ALLOW_ORIGIN=* \
    DISCOVERY_SERVICE="" \
    NUMBER_OF_MASTERS=1 \
    SSL_ENABLE=false \
    MODE=""

ADD config /elasticsearch/config

ADD fsloader /fsloader
RUN chmod +x /fsloader/*

RUN mkdir /elasticsearch/config/certs
RUN chown elasticsearch:elasticsearch -R /elasticsearch/config/certs

RUN mkdir /etc/service/fsloader
RUN ln -s /fsloader/run_fsloader.sh /etc/service/fsloader/run

RUN mkdir /etc/service/elasticsearch
RUN ln -s /run.sh /etc/service/elasticsearch/run

COPY yq /usr/bin/yq
COPY config-merger.sh /usr/bin/config-merger.sh
COPY runit.sh /runit.sh

ENTRYPOINT ["/runit.sh"]
