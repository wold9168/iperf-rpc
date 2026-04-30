FROM golang:1.22-alpine AS go-builder

RUN apk add --no-cache git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /iperf-rpc ./cmd/server

FROM node:22-alpine AS web-builder

RUN corepack enable && corepack prepare pnpm@latest --activate
WORKDIR /src
COPY web/package.json web/pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile
COPY web/ .
RUN pnpm run build

FROM alpine:3.20

RUN apk add --no-cache iperf3 ca-certificates tzdata

WORKDIR /app
COPY --from=go-builder /iperf-rpc .
COPY --from=web-builder /src/dist ./web/dist

EXPOSE 8080 5201

ENTRYPOINT ["/app/iperf-rpc"]
