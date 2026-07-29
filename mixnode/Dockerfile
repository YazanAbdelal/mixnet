FROM golang:1.26.5-alpine AS builder

WORKDIR /build

# the paths are relative to the project's root directory. this is becasue docker-compose sets the context as ./
COPY go.work go.sum* ./
COPY mixnode/go.mod mixnode/go.sum* ./mixnode/
COPY crypto/go.mod crypto/go.sum* ./crypto/
COPY proto/go.mod proto/go.sum* ./proto/
COPY keygen/go.mod keygen/go.sum* ./keygen/

# sync go.mod files accross modules
RUN go work sync

COPY mixnode/ ./mixnode/
COPY crypto/ ./crypto/
COPY proto/ ./proto/
COPY keygen/ ./keygen/

RUN go build -o /app/client ./mixnode/cmd
FROM alpine:latest

# `apk add` is used for downloading packages in Alpine Linux. 
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/client /app/client

ENTRYPOINT ["/app/client"]
