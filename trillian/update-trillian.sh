#!/bin/sh
# Script that updates the Trillian binaries in preparation for docker-compose

set -e

COMPOSE_DIR=$(pwd)
GOPATH=$(go env GOPATH)

# Fetch Trillian and its dependencies
go get -d -v -u github.com/google/trillian
cd $GOPATH/src/github.com/google/trillian
go get -d -v -t ./...
CGO_ENABLED=0 GOOS=linux go build ./...


cd $GOPATH/src/github.com/google/trillian
# Copy over the SQL initialization
cp storage/mysql/storage.sql $COMPOSE_DIR/db/01-init.sql
# Prepend a directive to use the appropriate database
sed -i '1iUSE trillian;' $COMPOSE_DIR/db/01-init.sql

cd $GOPATH/src/github.com/google/trillian
cd examples/ct/ct_server
CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o $COMPOSE_DIR/ct_server/main .

cd $GOPATH/src/github.com/google/trillian
cd server/trillian_log_signer
CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o $COMPOSE_DIR/log_signer/main .

cd $GOPATH/src/github.com/google/trillian
cd server/trillian_log_server
CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o $COMPOSE_DIR/log_server/main .

cd $COMPOSE_DIR/trampoline
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o trampoline trampoline.go
cp trampoline $COMPOSE_DIR/ct_server
cp trampoline $COMPOSE_DIR/log_signer
cp trampoline $COMPOSE_DIR/log_server
