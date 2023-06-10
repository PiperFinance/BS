FROM golang:1.20-alpine AS builder

RUN apk update && apk add alpine-sdk git && rm -rf /var/cache/apk/*

RUN mkdir -p /api
WORKDIR /api
ENV PORT=8000
COPY  ./go.mod .
COPY ./go.sum .
ENV GOPROXY=https://goproxy.cn,direct

RUN go mod download
COPY ./src ./src
RUN go build -o ./app  github.com/PiperFinance/BS/src

FROM alpine:latest

RUN apk update \
    && apk add ca-certificates  \
    && apk add --no-cache tzdata \
    && rm -rf /var/cache/apk/*

RUN mkdir -p /api
WORKDIR /api
COPY --from=builder /api/app .
COPY ./src/data/mainnets.json /api/data/mainnets.json
COPY ./src/data/mainnets.json /data/mainnets.json

RUN rm -rf /var/bs/log/ | true \ 
    && mkdir -p /var/bs/log/ \ 
    && touch /var/bs/log/err.log \ 
    && touch /var/bs/log/debug.log 

EXPOSE 7654

ENTRYPOINT /api/app
