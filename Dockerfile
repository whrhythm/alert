FROM golang:1.23 as builder

ADD . /src
WORKDIR /src
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -o app alertrules.go  client.go  config.go  contactpoints.go  main.go  notification.go

FROM alpine:latest
COPY --from=builder /src/app /usr/local/bin/app

EXPOSE 8080

CMD ["/usr/local/bin/app"]