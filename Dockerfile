FROM golang:alpine AS builder
WORKDIR /app
ADD . /app
RUN apk update && apk add build-base
RUN cd /app && GOARGS="-buildmode=pie" CGO_ENABLED=0 make orion -B

FROM scratch
WORKDIR /data
COPY --from=builder /app/orion /bin/orion
ENTRYPOINT ["/bin/orion", "-config", "/conf/orion.conf"]
VOLUME ["/conf"]
VOLUME ["/data"]
