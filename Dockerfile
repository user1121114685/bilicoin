# bilicoin service Dockerfile
# version 1.0.9
# author r3inbowari
FROM golang:1.17.2 as builder
LABEL stage="builder"

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY . .

# use netgo
ENV CGO_ENABLED=0

RUN chmod 777 build.sh
RUN  ./build.sh

RUN chmod 777 ./build/bilicoin_linux_amd64_v1.0.11

FROM alpine

WORKDIR /app

COPY --from=builder /app/build/bilicoin_linux_amd64_v1.0.11 .

ENTRYPOINT ["./bilicoin_linux_amd64_v1.0.11", "-a"]
