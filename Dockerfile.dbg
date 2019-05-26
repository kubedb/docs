FROM {ARG_FROM}

ENV DEBIAN_FRONTEND noninteractive
ENV DEBCONF_NONINTERACTIVE_SEEN true

RUN set -x \
  && apt-get update \
  && apt-get install -y --no-install-recommends apt-transport-https ca-certificates tzdata locales openssl \
  && rm -rf /var/lib/apt/lists/* /usr/share/doc /usr/share/man /tmp/* \
  && localedef -i en_US -c -f UTF-8 -A /usr/share/locale/locale.alias en_US.UTF-8 \
  && echo 'Etc/UTC' > /etc/timezone && dpkg-reconfigure tzdata

ENV TZ     :/etc/localtime
ENV LANG   en_US.utf8
ENV LC_ALL en_US.UTF-8

ADD bin/{ARG_OS}_{ARG_ARCH}/{ARG_BIN} /{ARG_BIN}

ENTRYPOINT ["/{ARG_BIN}"]
