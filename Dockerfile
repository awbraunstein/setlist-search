# Dockerfile References: https://docs.docker.com/engine/reference/builder/

FROM golang:1.12 as builder

ENV GO111MODULE=on
WORKDIR /go/src/github.com/awbraunstein/setlist-search

COPY . .

RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -o app .

FROM alpine

RUN apk --no-cache add ca-certificates

LABEL maintainer="Andrew Braunstein <awbraunstein@gmail.com>"

WORKDIR /root/
COPY --from=builder /go/src/github.com/awbraunstein/setlist-search/app .
COPY --from=builder /go/src/github.com/awbraunstein/setlist-search/templates templates/
COPY --from=builder /go/src/github.com/awbraunstein/setlist-search/.setsearcherindex .
ENV SETSEARCHERINDEX=/root/.setsearcherindex


EXPOSE 8080

CMD ["./app", "-http=:8080"]
