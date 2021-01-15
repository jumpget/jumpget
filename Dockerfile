## build container
FROM golang:alpine AS build-env

RUN apk update; apk add git build-base

WORKDIR $GOPATH/src/github.com/lsgrep/jumpget

ADD . $GOPATH/src/github.com/lsgrep/jumpget

# dep ensure
# RUN cd $GOPATH/src/github.com/dexDev/dex3; go get -u github.com/golang/dep/cmd/dep; dep ensure

RUN go build -o app -a -ldflags '-extldflags "-static"' github.com/lsgrep/jumpget; mv app /app

## final container
FROM alpine
RUN apk update && apk add ca-certificates bash && rm -rf /var/cache/apk/*

VOLUME /data
RUN addgroup -S jumpget && adduser -S jumpget -G jumpget
RUN chown -R jumpget /data
USER jumpget

WORKDIR /work
#RUN mkdir /data

COPY --from=build-env /app /work/app
EXPOSE 3100 4100

CMD /work/app --server
