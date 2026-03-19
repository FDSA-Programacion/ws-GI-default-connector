FROM golang:1.25 AS builder

ARG SKIP_INTEGRATION_TESTS=0
ENV SKIP_INTEGRATION_TESTS=${SKIP_INTEGRATION_TESTS}

COPY . /go/src/application

WORKDIR /go/src/application/app

RUN go mod tidy
RUN go mod verify

RUN go test ./...

WORKDIR /go/src/application/app/cmd/api
RUN CGO_ENABLED=1 go build -o binary

FROM ubuntu:24.04

ENV TZ=Europe/Madrid

# Install CA certificates and other dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    gcc \
    libc6-dev \
    libaio1t64 \
    wget \
    unzip \
    gettext-base \
    && rm -rf /var/lib/apt/lists/*

# Setup the oracle driver...
RUN apt-get update && apt-get install -y --no-install-recommends gcc libc6-dev libaio1t64 wget unzip gettext-base
RUN ln -s /usr/lib/x86_64-linux-gnu/libaio.so.1t64 /usr/lib/x86_64-linux-gnu/libaio.so.1
RUN wget --no-check-certificate -O /tmp/instantclient-basic-linux-x64.zip https://download.oracle.com/otn_software/linux/instantclient/193000/instantclient-basic-linux.x64-19.3.0.0.0dbru.zip
RUN mkdir -p /usr/lib/oracle && unzip /tmp/instantclient-basic-linux-x64.zip -d /usr/lib/oracle
RUN ldconfig -v /usr/lib/oracle/instantclient_19_3
RUN ldd /usr/lib/oracle/instantclient_19_3/libclntsh.so

RUN mkdir -p /application/app
WORKDIR /application/app/cmd/api

COPY assets /application/assets
COPY --from=0 /go/src/application/app/cmd/api/binary .

RUN chmod +x /application/assets/run.sh


CMD ["/application/assets/run.sh"]
