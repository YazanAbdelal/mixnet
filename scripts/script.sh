#!/bin/bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# get number of mix servers and clients in the mixnet
read -p "Number of mix nodes: " num_nodes
read -p "Number of clients: " num_clients

# run the keygen binary from the root directory to output to the ./keys folder
go run ./keygen/cmd -servers "$num_nodes" -clients "$num_clients"

# generate CA certificate
mkdir -p certs
openssl genrsa -out certs/ca.key 2048
openssl req -x509 -new -nodes -key certs/ca.key -sha256 -days 365 -out certs/ca.crt -subj "/CN=mixnet-CA" 

# generate TLS certificates for all nodes (mTLS)
for ((i = 1; i <= num_nodes; i++)); do
  name="mixnode-$i"
  openssl genrsa -out "certs/$name.key" 2048
  openssl req -new -key "certs/$name.key" -out "certs/$name.csr" -subj "/CN=$name" -addext "subjectAltName=DNS:$name,DNS:$name" # -addext is for SAN
  openssl x509 -req -in "certs/$name.csr" -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -out "certs/$name.crt" -days 365 -sha256 -copy_extensions copy # -copy_extentions copies the SAN
  rm "certs/$name.csr"
done

for ((i = 1; i <= num_clients; i++)); do
  name="client-$i"
  openssl genrsa -out "certs/$name.key" 2048
  openssl req -new -key "certs/$name.key" -out "certs/$name.csr" -subj "/CN=$name" -addext "subjectAltName=DNS:$name,DNS:$name"
  openssl x509 -req -in "certs/$name.csr" -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -out "certs/$name.crt" -days 365 -sha256 -copy_extensions copy
  rm "certs/$name.csr"
done

# generate docker-compose.yml based on the requested topology
cat > docker-compose.yml <<COMPOSE_EOF

networks:
  mixnet:

services:
COMPOSE_EOF

for ((i = 1; i <= num_nodes; i++)); do
  cat >> docker-compose.yml <<COMPOSE_EOF
  mixnode-$i:
    container_name: server-${i}
    build:
      context: .
      dockerfile: server/Dockerfile
    volumes:
      - ./keys/server_${i}_private.pem:/etc/mixnet/keys/private.pem:ro
      - ./keys/public:/etc/mixnet/keys/public:ro
      - ./certs/mixnode-${i}.crt:/etc/mixnet/certs/tls.crt:ro
      - ./certs/mixnode-${i}.key:/etc/mixnet/certs/tls.key:ro
      - ./certs/ca.crt:/etc/mixnet/certs/ca.crt:ro
    ports:
      - "500${i}:500${i}"
    networks:
      - mixnet

COMPOSE_EOF
done

for ((i = 1; i <= num_clients; i++)); do
  clientPort=$((50050 + i))
  cat >> docker-compose.yml <<COMPOSE_EOF
  client-$i:
    container_name: client-${i}
    build:
      context: .
      dockerfile: client/Dockerfile
    command: ["--port", "$clientPort"]
    stdin_open: true
    tty: true
    volumes:
      - ./keys/client_${i}_private.pem:/etc/mixnet/keys/private.pem:ro
      - ./keys/public:/etc/mixnet/keys/public:ro
      - ./certs/client-${i}.crt:/etc/mixnet/certs/tls.crt:ro
      - ./certs/client-${i}.key:/etc/mixnet/certs/tls.key:ro
      - ./certs/ca.crt:/etc/mixnet/certs/ca.crt:ro
    depends_on:
COMPOSE_EOF
  for ((j = 1; j <= num_nodes; j++)); do
    echo "      - mixnode-$j" >> docker-compose.yml
  done
  cat >> docker-compose.yml <<COMPOSE_EOF
    networks:
      - mixnet

COMPOSE_EOF
done