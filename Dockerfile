## build container
FROM golang:alpine AS build-env

RUN apk update; apk add git build-base

WORKDIR $GOPATH/src/github.com/lsgrep/jumpget

ADD . $GOPATH/src/github.com/lsgrep/jumpget


RUN go build -o app -a -ldflags '-extldflags "-static"' github.com/lsgrep/jumpget; mv app /app

## final container
FROM alpine
RUN apk update && apk add ca-certificates bash && rm -rf /var/cache/apk/*
RUN addgroup -S jumpget && adduser -S -G jumpget -h /home/jumpget jumpget
RUN mkdir /home/jumpget; chown -R jumpget:jumpget /home/jumpget
USER jumpget


RUN mkdir -p /home/jumpget/data
VOLUME /home/jumpget/data
ENV JUMPGET_DATA_DIR /home/jumpget/data

WORKDIR /work
#RUN mkdir /data

COPY --from=build-env /app /work/app
EXPOSE 3100 4100

CMD /work/app --server
