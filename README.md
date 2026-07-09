# mixnet
A mixnet implementation in Go using Docker compose.

## How to Start the Mixnet:
Run the script: `./script.sh` and input the number of mixnet servers and clients. This will create the keys and docker-compose.yml file.
Then run the mixnet using Docker compose: `docker compose up -d`

## For mTLS Encryption:
#### First we create a **Certificate Authority** (CA):
generate a RSA key of length 2048 and save it in ca.key

this key is private and is used to sign cerficates.

``` sh
openssl genrsa -out ca.key 2048
```
___
generate a self-signed (x509) certificate using the private key in ca.key, and save the certificate in ca.crt. And set the identity of the CA as mixnet-CA.

`-nodes` means that the cerificate is not password-protected.

this certificate is public and is used to verify signatures.

```sh
openssl req -x509 -new -nodes -key ca.key -sha256 -days 365 -out ca.crt -subj "/CN=mixnet-CA" 
```
___
#### Then we create a certificate for each of the nodes:
>here NODE_NAME should match the Docker service name.

generate a new RSA key.
```sh
openssl genrsa -out NODE_NAME.key 2048
```
___
generate a new **Certificate Signing Request (CSR)**: The node sends its public key (derived from the private key) and its identity and signs them using its private key.
```sh
openssl req -new -key NODE_NAME.key -out NODE_NAME.csr -subj "/CN=NODE_NAME"
```
___
CA signs certificate using its private key: Takes in a CSR as input (`-req`), specifies the certificate holder (`-in NODE_NAME.csr`), and the Certificate Authority (CA) (`-CAkey ca.key`) and creates a serial number (not important) (`-CAcreateserial`).
```sh
openssl x509 -req -in NODE_NAME.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out NODE_NAME.crt -days 365 -sha256
```
___

## Generating gRPC code:
First install protobuf compiler.
Then install the go packages:
``` sh
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```
then export the go binaries to PATH:
```sh 
export PATH="$PATH:$(go env GOPATH)/bin"
```
then run the following commands:
```sh
cd proto
protoc --go_out=paths=source_relative:gen --go-grpc_out=paths=source_relative:gen mixnet.proto
```