FROM golang:1.20-alpine AS builder

RUN apk update && apk add alpine-sdk git && rm -rf /var/cache/apk/*

RUN mkdir -p /api
WORKDIR /api
ENV PORT=8000
COPY  ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY ./src ./src
COPY ./data ./data
RUN go build -o ./app  github.com/PiperFinance/BS/src

FROM alpine:latest

RUN apk update \
    && apk add ca-certificates  \
    && apk add --no-cache tzdata \
    && rm -rf /var/cache/apk/* 

RUN mkdir -p /api
WORKDIR /api
COPY --from=builder /api/app .
COPY ./data /api/data
# COPY server-config /api/config
EXPOSE 7654

ENTRYPOINT env && /api/app
