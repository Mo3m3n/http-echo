FROM golang:1.17-alpine AS builder

COPY . /src/

RUN cd /src && go build -o echo-http


FROM alpine:latest

WORKDIR /app

COPY --from=builder /src/echo-http .

COPY generate-cert.sh .

RUN apk --no-cache add openssl

ENTRYPOINT ["./echo-http"]
CMD []
