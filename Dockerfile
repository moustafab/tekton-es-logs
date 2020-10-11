FROM golang:1.14 as builder

ARG VERSION=snapshot
WORKDIR /go/src/app
COPY . .

RUN CGO_ENABLED=0 make VERSION=$VERSION RELEASE=1 build

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/app/build/tekton-es-logs .

ENTRYPOINT ["./tekton-es-logs"]
EXPOSE 8080
ENV ELASTICSEARCH_URL="es.domain.com"


