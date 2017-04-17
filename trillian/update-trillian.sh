#!/bin/bash
# Script that updates the Trillian binaries in preparation for docker-compose

set -e

COMPOSE_DIR=$(pwd)
GOPATH=$(go env GOPATH)

if [[ -e $GOPATH/src/github.com/google/trillian ]]; then
    cd $GOPATH/src/github.com/google/trillian
    git fetch
    if [[ "$(git rev-parse HEAD)" != "$(git rev-parse origin)" ]]; then
        # Fetch Trillian and its dependencies
        printf "Updating Trillian\n"
        go get -d -u github.com/google/trillian
        cd $GOPATH/src/github.com/google/trillian
        go get -d -t ./...
    else
        printf "Skipping Trillian update because Trillian is already up to date.\n"
    fi
else
    # Fetch Trillian and its dependencies
    printf "Downloading Trillian\n"
    go get -d -u github.com/google/trillian
    cd $GOPATH/src/github.com/google/trillian
    go get -d -t ./...
fi


cd $GOPATH/src/github.com/google/trillian
# CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo ./...

# Copy over the SQL initialization
cp storage/mysql/storage.sql $COMPOSE_DIR/db/01-init.sql
# Prepend a directive to use the appropriate database
sed -i '1iUSE trillian;' $COMPOSE_DIR/db/01-init.sql

# patch Trillian
cd $GOPATH/src/github.com/google/trillian/server/trillian_log_server
printf "Patching trillian\n"
sed -i s/localhost/0.0.0.0/ main.go
cd $GOPATH/src/github.com/google/trillian/server

printf "Building trillian\n"
cd $COMPOSE_DIR
CGO_ENABLED=0 GOOS=linux go install -a -installsuffix cgo github.com/google/trillian/server/trillian_log_server github.com/google/trillian/server/trillian_log_signer github.com/google/trillian/examples/ct/ct_server github.com/google/trillian/cmd/createtree
mv $GOPATH/bin/createtree ct_server/
mv $GOPATH/bin/ct_server ct_server/main
mv $GOPATH/bin/trillian_log_server log_server/main
mv $GOPATH/bin/trillian_log_signer log_signer/main

cd $COMPOSE_DIR/trampoline
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o trampoline trampoline.go
cp trampoline $COMPOSE_DIR/ct_server
cp trampoline $COMPOSE_DIR/log_signer
cp trampoline $COMPOSE_DIR/log_server

cd $COMPOSE_DIR/ct_server
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o launcher launcher.go
