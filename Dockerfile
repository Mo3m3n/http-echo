FROM golang:1.14.2-alpine AS builder

COPY main.go /src/

RUN cd /src && go build -o echo-http


FROM alpine:latest

WORKDIR /app

COPY --from=builder /src/echo-http .

COPY generate-cert.sh .

ENV HTTP_PORT=80 HTTPS_PORT=443

RUN apk --no-cache add openssl

ENTRYPOINT ["./echo-http"]
CMD []
