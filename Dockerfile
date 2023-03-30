FROM golang:1.20.2-alpine3.17 as builder
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY ./main.go .
RUN go build -v -o /usr/local/bin/app ./...

FROM alpine:3.17
ARG TZ="Asia/Shanghai"
ENV TZ ${TZ}
RUN apk add --no-cache bash tzdata ca-certificates \
    && ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone
COPY --from=builder /usr/local/bin/app /root/app
WORKDIR /
CMD ["/root/app"]
