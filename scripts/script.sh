#!/bin/bash
set -euo pipefail

generate_ca_cert() {
	mkdir -p certs
	openssl genrsa -out certs/ca.key 2048
	openssl req -x509 -new -nodes -key certs/ca.key -sha256 -days 365 \
		-out certs/ca.crt -subj "/CN=mixnet-CA"
}

generate_tls_cert() {
	local name="$1"

	openssl genrsa -out "certs/$name.key" 2048
	openssl req -new -key "certs/$name.key" -out "certs/$name.csr" \
		-subj "/CN=$name" \
		-addext "subjectAltName=DNS:$name"
	openssl x509 -req -in "certs/$name.csr" \
		-CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial \
		-out "certs/$name.crt" -days 365 -sha256 -copy_extensions copy
	rm "certs/$name.csr"
}

generate_tls_certs() {
	generate_ca_cert

	for ((i = 1; i <= num_nodes; i++)); do
		generate_tls_cert "server-$i"
	done

	for ((i = 1; i <= num_clients; i++)); do
		generate_tls_cert "client-$i"
	done
}

generate_compose_file() {
	cat > docker-compose.yml <<COMPOSE_EOF
networks:
  mixnet:

services:
COMPOSE_EOF

	for ((i = 1; i <= num_nodes; i++)); do
		cat >> docker-compose.yml <<COMPOSE_EOF
  server-$i:
    container_name: server-${i}
    build:
      context: .
      dockerfile: mixnode/Dockerfile
    command: ["--type", "server", "--name", "server-$i"]
    volumes:
      - ./keys/server-${i}-private.pem:/etc/mixnet/keys/private.pem:ro
      - ./keys/public:/etc/mixnet/keys/public:ro
      - ./certs/server-${i}.crt:/etc/mixnet/certs/tls.crt:ro
      - ./certs/server-${i}.key:/etc/mixnet/certs/tls.key:ro
      - ./certs/ca.crt:/etc/mixnet/certs/ca.crt:ro
    networks:
      - mixnet

COMPOSE_EOF
	done

	for ((i = 1; i <= num_clients; i++)); do
		cat >> docker-compose.yml <<COMPOSE_EOF
  client-$i:
    container_name: client-${i}
    build:
      context: .
      dockerfile: mixnode/Dockerfile
    command: ["--type", "client", "--name", "client-$i", "--dest", ""]
    stdin_open: true
    tty: true
    volumes:
      - ./keys/client-${i}-private.pem:/etc/mixnet/keys/private.pem:ro
      - ./keys/public:/etc/mixnet/keys/public:ro
      - ./certs/client-${i}.crt:/etc/mixnet/certs/tls.crt:ro
      - ./certs/client-${i}.key:/etc/mixnet/certs/tls.key:ro
      - ./certs/ca.crt:/etc/mixnet/certs/ca.crt:ro
    depends_on:
COMPOSE_EOF
		for ((j = 1; j <= num_nodes; j++)); do
			echo "      - server-$j" >> docker-compose.yml
		done
		cat >> docker-compose.yml <<COMPOSE_EOF
    networks:
      - mixnet

COMPOSE_EOF
	done
}

# main flow
read -p "Number of mix nodes: " num_nodes
read -p "Number of clients: " num_clients

go run ./keygen/cmd -servers "$num_nodes" -clients "$num_clients"

generate_tls_certs
generate_compose_file
