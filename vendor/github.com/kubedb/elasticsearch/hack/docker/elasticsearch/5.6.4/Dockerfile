FROM quay.io/pires/docker-elasticsearch:5.6.4

RUN set -x \
	&& apk add --update --no-cache runit curl

ENV NODE_NAME=""

# Install mapper-attachments (https://www.elastic.co/guide/en/elasticsearch/plugins/5.x/mapper-attachments.html)
RUN ./bin/elasticsearch-plugin install ingest-attachment

# Install search-guard
RUN ./bin/elasticsearch-plugin install -b com.floragunn:search-guard-5:5.6.4-19.1

RUN chmod +x -R plugins/search-guard-5/tools/*.sh

# Set environment variables defaults
ENV ES_JAVA_OPTS "-Xms512m -Xmx512m"
ENV CLUSTER_NAME "elasticsearch"
ENV NODE_MASTER true
ENV NODE_DATA true
ENV NODE_INGEST true
ENV HTTP_ENABLE true
ENV HTTP_CORS_ENABLE true
ENV HTTP_CORS_ALLOW_ORIGIN *
ENV DISCOVERY_SERVICE ""
ENV NUMBER_OF_MASTERS 1
ENV SSL_ENABLE=false
ENV MODE=""

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
