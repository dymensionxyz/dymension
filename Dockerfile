FROM golang:1.19-alpine3.16 as go-builder

WORKDIR /app

COPY go.mod go.sum* ./

RUN go mod download

COPY . .

ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3

RUN apk add --no-cache $PACKAGES

RUN make build

FROM alpine:3.16.1

RUN apk add curl jq bash vim 

COPY --from=go-builder /app/build/dymd /usr/local/bin/

WORKDIR /app

COPY scripts/* ./scripts/

ENV CHAIN_ID=local-testnet
ENV KEY_NAME=local-user
ENV MONIKER_NAME=local

RUN chmod +x ./scripts/*.sh

EXPOSE 26656 26657 1317 9090
