#!/bin/bash
# mixnet setup script
# Generates RSA keys, mTLS certificates, node config, and docker-compose.yml
# based on the user-specified number of mix servers and clients.
set -euo pipefail

# -------------------------------------------------------
#  1. CREATE A CERTIFICATE AUTHORITY (CA)
# -------------------------------------------------------
# The CA is a self-signed certificate used to sign individual
# node certificates. Every node trusts the CA, which allows
# mutual TLS authentication between any pair of nodes.
generate_ca_cert() {
	mkdir -p certs

	# generate a 2048-bit RSA key for the CA (private, kept secret)
	openssl genrsa -out certs/ca.key 2048

	# create a self-signed CA certificate (public, distributed to all nodes)
	openssl req -x509 -new -nodes -key certs/ca.key -sha256 -days 365 \
		-out certs/ca.crt -subj "/CN=mixnet-CA"
}

# -------------------------------------------------------
#  2. CREATE A TLS CERTIFICATE FOR A SINGLE NODE
# -------------------------------------------------------
# Each node (server or client) gets its own RSA key pair and a
# certificate signed by the CA. These are used for mTLS, every
# gRPC connection between nodes is mutually authenticated.
generate_tls_cert() {
	local name="$1"

	# generate the node's private RSA key
	openssl genrsa -out "certs/$name.key" 2048

	# create a Certificate Signing Request (CSR) containing the
	# node's public key and identity. The SAN (Subject Alternative Name)
	# is set to the node's Docker service name so that TLS hostname
	# verification succeeds inside the Docker network.
	openssl req -new -key "certs/$name.key" -out "certs/$name.csr" \
		-subj "/CN=$name" \
		-addext "subjectAltName=DNS:$name"

	# the CA signs the CSR, producing the node's certificate
	openssl x509 -req -in "certs/$name.csr" \
		-CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial \
		-out "certs/$name.crt" -days 365 -sha256 -copy_extensions copy

	rm "certs/$name.csr"
}

# -------------------------------------------------------
#  3. GENERATE CERTIFICATES FOR ALL NODES
# -------------------------------------------------------
# First create the CA, then a certificate for every server
# and every client.
generate_tls_certs() {
	generate_ca_cert

	for ((i = 1; i <= num_nodes; i++)); do
		generate_tls_cert "server-$i"
	done

	for ((i = 1; i <= num_clients; i++)); do
		generate_tls_cert "client-$i"
	done
}

# -------------------------------------------------------
#  4. GENERATE NODE CONFIGURATION FILE
# -------------------------------------------------------
# Produces config.json listing all servers, clients, and tunable parameters.
# The application loads this file at startup, no rebuild needed to
# change batch sizes, network timing, or path length.
generate_config() {
	servers_json=""
	for ((i = 1; i <= num_nodes; i++)); do
		[ -n "$servers_json" ] && servers_json+=", "
		servers_json+="\"server-$i\""
	done

	clients_json=""
	for ((i = 1; i <= num_clients; i++)); do
		[ -n "$clients_json" ] && clients_json+=", "
		clients_json+="\"client-$i\""
	done

	cat > config.json <<EOF
{
  "servers": [$servers_json],
  "clients": [$clients_json],
  "path_len": $path_len,
  "batch_size": $batch_size,
  "client_tick_us": $client_tick_us,
  "flush_timeout_ms": $flush_timeout_ms
}
EOF
}

# -------------------------------------------------------
#  5. GENERATE DOCKER COMPOSE FILE
# -------------------------------------------------------
# Creates a docker-compose.yml with one service per node.
# Each service:
#   - builds from the same Dockerfile (mixnode/Dockerfile)
#   - receives its type, name and (for clients) destination
#     as command-line flags
#   - mounts its private key, public keys directory, TLS
#     certificate/key, CA cert, and config.json as read-only
#     volumes at the paths the application expects
#   - clients have stdin_open + tty so the user can type
#     messages via `docker attach`
#   - clients depend_on all servers so they start after
#     the mix network is ready
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
      - ./config.json:/etc/mixnet/config.json:ro
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
      - ./config.json:/etc/mixnet/config.json:ro
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

# -------------------------------------------------------
#  MAIN FLOW
# -------------------------------------------------------

# ask the user for the network size
read -p "Number of mix nodes: " num_nodes
read -p "Number of clients: " num_clients
read -p "Path length (number of mix hops, default 3): " path_len
path_len=${path_len:-3}
read -p "Batch size (messages per mix round, default 10): " batch_size
batch_size=${batch_size:-10}
read -p "Client tick (microseconds between sends, default 200): " client_tick_us
client_tick_us=${client_tick_us:-200000}
read -p "Flush timeout (ms before a partial batch is sent, default 500): " flush_timeout_ms
flush_timeout_ms=${flush_timeout_ms:-1}

# step 1: generate RSA key pairs (private + public) for each node
#         (used for onion encryption, not TLS)
go run ./keygen/cmd -servers "$num_nodes" -clients "$num_clients"

# step 2: generate mTLS certificates (CA + one per node)
#         (used for authenticated gRPC channels)
generate_tls_certs

# step 3: write config.json listing all servers and clients
generate_config

# step 4: write docker-compose.yml with all services, volumes and networks
generate_compose_file
