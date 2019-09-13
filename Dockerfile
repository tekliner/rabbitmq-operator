FROM golang:1.13-alpine

RUN apk update
RUN apk add git mercurial

WORKDIR /app/rabbitmq-operator
COPY . .

WORKDIR /app/rabbitmq-operator/cmd/manager

RUN CGO_ENABLED=0 go build -mod=vendor
RUN go install -mod=vendor

FROM alpine:3.10

COPY --from=0 /go/bin/manager /usr/local/bin/rabbitmq-operator

ENTRYPOINT ["/usr/local/bin/entrypoint"]
