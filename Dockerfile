FROM golang:1.18.4-alpine3.16 as go-builder

WORKDIR /app

COPY go.mod go.sum* ./

RUN go mod download

COPY . .

ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3

RUN apk add --no-cache $PACKAGES

RUN make build

FROM alpine:3.16.1

RUN apk add curl jq

COPY --from=go-builder /app/bin/dymd /usr/local/bin/

COPY scripts/* /scripts/

RUN chmod +x /scripts/*.sh

EXPOSE 26656 26657 1317 9090

CMD ["/bin/sh"]
