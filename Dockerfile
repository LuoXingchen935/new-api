FROM oven/bun:latest AS builder

WORKDIR /build
# 首先复制 VERSION 文件
COPY VERSION .
# 然后复制前端项目相关文件
COPY web/package.json web/bun.lock ./
RUN bun install
COPY web/ .
# 设置构建环境变量并执行构建
RUN VITE_REACT_APP_VERSION=$(cat VERSION) DISABLE_ESLINT_PLUGIN=true bun run build

FROM golang:alpine AS builder2

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /build

ADD go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=builder /build/dist ./web/dist
RUN VERSION=$(cat VERSION) && go build -ldflags "-s -w -X 'one-api/common.Version=$VERSION'" -o one-api

FROM alpine

RUN apk upgrade --no-cache \
    && apk add --no-cache ca-certificates tzdata ffmpeg \
    && update-ca-certificates

COPY --from=builder2 /build/one-api /
EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/one-api"]
