# run: make image
FROM golang:1.14

WORKDIR /build
ADD . /build

RUN git submodule update --init --recursive -j 8

RUN make init

ENV GOOS=linux
ENV GOARCH=386
RUN make build

FROM asciidoctor/docker-asciidoctor:1.1.0

RUN apk add --no-cache \
    git

COPY --from=0 build/monako /usr/bin/monako
RUN chmod +x /usr/bin/monako

WORKDIR /docs
