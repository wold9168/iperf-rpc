FROM alpine:3.20

RUN apk add --no-cache iperf3 ca-certificates tzdata

WORKDIR /app
COPY iperf-rpc .
COPY web/dist ./web/dist

EXPOSE 8080 5201

ENTRYPOINT ["/app/iperf-rpc"]
