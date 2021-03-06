# Dockerfile References: https://docs.docker.com/engine/reference/builder/

FROM golang:1.12 as builder

ENV GO111MODULE=on
WORKDIR /go/src/github.com/awbraunstein/setlist-search

COPY . .

RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -o app .

FROM alpine

RUN apk update \
        && apk upgrade \
        && apk add --no-cache \
        ca-certificates \
        && update-ca-certificates 2>/dev/null || true

LABEL maintainer="Andrew Braunstein <awbraunstein@gmail.com>"

WORKDIR /root/
COPY --from=builder /go/src/github.com/awbraunstein/setlist-search/app .
COPY --from=builder /go/src/github.com/awbraunstein/setlist-search/templates templates/
COPY --from=builder /go/src/github.com/awbraunstein/setlist-search/assets assets/

EXPOSE 8080

CMD ["./app"]
