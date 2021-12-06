# Builder image
FROM golang:1.17-alpine AS builder
WORKDIR /go/src/
RUN apk add git \
&& git clone https://github.com/genecywang/prom2mdi.git \
&& cd prom2mdi \
&& CGO_ENABLED=0 go build -o prom2mdi


# Final Stage
FROM gliderlabs/alpine:latest
WORKDIR /usr/local/bin/
COPY --from=builder /go/src/prom2mdi/prom2mdi /usr/local/bin/
COPY --from=builder /go/src/prom2mdi/configs /configs
EXPOSE 9000
ENTRYPOINT ["prom2mdi","--config","/configs/config.yaml","--template","/configs/mdi.tmpl"]