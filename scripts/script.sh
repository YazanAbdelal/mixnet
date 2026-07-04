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
  openssl req -new -key "certs/$name.key" -out "certs/$name.csr" -subj "/CN=$name"
  openssl x509 -req -in "certs/$name.csr" -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -out "certs/$name.crt" -days 365 -sha256
  rm "certs/$name.csr"
done

for ((i = 1; i <= num_clients; i++)); do
  name="client-$i"
  openssl genrsa -out "certs/$name.key" 2048
  openssl req -new -key "certs/$name.key" -out "certs/$name.csr" -subj "/CN=$name"
  openssl x509 -req -in "certs/$name.csr" -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -out "certs/$name.crt" -days 365 -sha256
  rm "certs/$name.csr"
done

# generate docker-compose.yml based on the requested topology
cat > "$script_dir/docker-compose.yml" <<COMPOSE_EOF
version: '3.8'

networks:
  mixnet:

services:
COMPOSE_EOF

for ((i = 1; i <= num_nodes; i++)); do
  cat >> "$script_dir/docker-compose.yml" <<COMPOSE_EOF
  mixnode-$i:
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
  cat >> "$script_dir/docker-compose.yml" <<COMPOSE_EOF
  client-$i:
    build:
      context: .
      dockerfile: client/Dockerfile
    volumes:
      - ./keys/client_${i}_private.pem:/etc/mixnet/keys/private.pem:ro
      - ./keys/public:/etc/mixnet/keys/public:ro
      - ./certs/client-${i}.crt:/etc/mixnet/certs/tls.crt:ro
      - ./certs/client-${i}.key:/etc/mixnet/certs/tls.key:ro
      - ./certs/ca.crt:/etc/mixnet/certs/ca.crt:ro
    depends_on:
    networks:
      - mixnet
COMPOSE_EOF
  for ((j = 1; j <= num_nodes; j++)); do
    echo "      - mixnode-$j" >> "$script_dir/docker-compose.yml"
  done
  echo "" >> "$script_dir/docker-compose.yml"
done