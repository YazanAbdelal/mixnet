FROM golang:1.25-alpine AS builder

WORKDIR /build

# the paths are relative to the project's root directory. this is becasue docker-compose sets the context as ./
COPY go.work go.sum* ./
COPY client/go.mod client/go.sum* ./client/
COPY crypto/go.mod crypto/go.sum* ./crypto/
COPY proto/go.mod proto/go.sum* ./proto/
COPY keygen/go.mod keygen/go.sum* ./keygen/

# sync go.mod files accross modules
RUN go work sync

COPY client/ ./client/
COPY crypto/ ./crypto/
COPY proto/ ./proto/
COPY keygen/ ./keygen/

RUN go build -o /app/client ./client/cmd
FROM alpine:latest

# `apk add` is used for downloading packages in Alpine Linux. 
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/client /app/client

ENTRYPOINT ["/app/client"]
