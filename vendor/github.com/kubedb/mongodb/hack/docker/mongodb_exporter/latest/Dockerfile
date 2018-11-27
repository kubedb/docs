# ref: https://github.com/dcu/mongodb_exporter/blob/master/Dockerfile

FROM golang:alpine as builder

RUN apk --no-cache add curl git make perl
RUN curl -s https://glide.sh/get | sh

RUN git clone https://github.com/dcu/mongodb_exporter.git /go/src/github.com/dcu/mongodb_exporter \
    && cd /go/src/github.com/dcu/mongodb_exporter \
    && git checkout master \
    && make release

FROM       alpine
EXPOSE     9001

RUN apk add --update ca-certificates
COPY --from=builder /go/src/github.com/dcu/mongodb_exporter/release/mongodb_exporter-linux-amd64 /usr/local/bin/mongodb_exporter

ENTRYPOINT [ "mongodb_exporter" ]
