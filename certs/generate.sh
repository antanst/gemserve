#!/bin/sh
set -eu

# Generate private key for CA
openssl genrsa -out ca.key 4096

# Generate CA certificate
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.crt \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=My CA"

# Generate private key for server
openssl genrsa -out server.key 2048

# Generate Certificate Signing Request (CSR) for server
openssl req -new -key server.key -out server.csr \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"

# Generate server certificate signed by our CA
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key \
	-CAcreateserial -out server.crt -days 3650 -sha256
