# bilicoin service Dockerfile
# version 1.0.6
# author r3inbowari
FROM golang:1.14 as builder
LABEL stage="builder"

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY . .

# 使用 netgo
ENV CGO_ENABLED=0

RUN go build cmd/main.go
RUN chmod 777 main

FROM alpine

WORKDIR /app

COPY --from=builder /app/main .

ENTRYPOINT ["./main", "-a"]
