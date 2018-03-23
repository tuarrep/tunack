FROM golang:1.10 as builder

COPY ./*.go /go/src/dahus.io/tunack/

WORKDIR /go/src/dahus.io/tunack

RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go install -v ./...

CMD ["/go/bin/tunack"]

FROM alpine

WORKDIR /app

COPY --from=builder /go/bin/tunack .

ENTRYPOINT  ["/bin/sh", "-c", "./tunack --inCluster"]