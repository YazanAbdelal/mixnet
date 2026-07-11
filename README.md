# mixnet
A mixnet implementation in Go using Docker compose.

## How to Start the Mixnet:
Run the script: `./scripts/script.sh` and input the number of mixnet servers and clients. This will create the keys, certs and docker-compose.yml file.
In the generated `docker-compose.yml` file, the destinations that each client sends to have to be specified by the user.
Then run the mixnet using Docker compose: `docker compose up -d --build`
In order to send messages, run the following command: `docker attach [CONTAINER_NAME]`, where `CONTAINER_NAME` is the name of the sender node.

## For mTLS Encryption:
*The following is done automatically by the script when we run ./scripts/script.sh*
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
___
# How the Mixnet Works:
**Onion Encryption:**

Each layer is encrypted as follows:
1. First we generate an ephemeral AES key of length 32 Bytes.
2. Then we use the AES key to encrypt the ciphertext we received from the previous layer (or the plaintext in case this is the innermost layer) - call output of this step 'c'.
3. We then calculate l = len(c).
4. We then encrypt both the AES key, l and the next node's name using the public RSA key of the next node.
5. The final output should be: [encrypted (AES key, l)][c], where the length of the first part is always 512B because we are using a 4096 bit (=512B) RSA key. 

When we are done adding all the layers, we pad the message with random bits up to 4096 Bytes.

Each layer is decrypted as follows:
1. First we decrypt the first 512 Bytes of the packet using the current node's private key, and we get: AES key, l.
2. We use l to separate the ciphertext from the padding, and we get c.
3. We use the AES key to decrypt c, and get the content, and the next node.
4. If the current node is the destination, we are done. Otherwise, we pad the content, and forward to the next node.

**Sending Messages:**

A message is sent from Client1 to Client2 as follows:
1. Client1 receives the message and from stdin through the receive only channel `input`.
2. The sendLoop function reads the message from `input`, encrypts it using the `OnionEncrypt` function in the `crypto` module. Then it creates a stub to the first server in the mixnode and calls `ForwardMessage`.
3. Each mix-node consecutively receieves the message from the previous node, decrypts a layer, and then forwards to the next layer until the message reaches the destination, who then prints or stores it.

**Receiving Messages:**

Every client starts a listener on port 50050, and assigns a gRPC handler to handle incoming gRPC calls through a mTLS-encrypted channel:
1. When a packet arrives at port 50050, the gRPC server handles it and calls the `ForwardMessage` method.
2. `ForwardMessage` then hands the packet to the `receiveMessages` function using the `Pipe` channel.
3. `receiveMessages` decrypts the outer layer , pads the message until it is of a preset length, and forwards it to the next Message.
4. Steps 1, 2, 3 are repeated until the message arrives to its final destination, the last node in the path, only this time the message is not padded and forwarded, but printed or stored.