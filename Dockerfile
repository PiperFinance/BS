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
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -o ./app  

FROM alpine:latest

RUN apk update \
    && apk add ca-certificates  \
    && apk add --no-cache tzdata \
    && rm -rf /var/cache/apk/*

RUN mkdir -p /api
WORKDIR /api
COPY --from=builder /api/app .
ADD https://raw.githubusercontent.com/PiperFinance/CD/main/chains/mainnetV2.json /data/mainnets.json  

RUN rm -rf /var/bs/log/ | true \ 
    && mkdir -p /var/bs/log/ \ 
    && touch /var/bs/log/err.log \ 
    && touch /var/bs/log/debug.log 

EXPOSE 7654

ENTRYPOINT /api/app
