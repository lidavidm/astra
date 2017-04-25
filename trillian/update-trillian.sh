#!/bin/bash
# Script that updates the Trillian binaries in preparation for docker-compose

set -e

TRILLIAN_VERSION="486b2c10b9e62cfa12ee93e50fdc23d0288c3b19"
COMPOSE_DIR=$(pwd)
GOPATH=$(go env GOPATH)

if [[ -e $GOPATH/src/github.com/google/trillian ]]; then
    cd $GOPATH/src/github.com/google/trillian
    git checkout master
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

printf "Checking out Trillian version $TRILLIAN_VERSION\n"
git checkout $TRILLIAN_VERSION

compile() {
    # Allow caller to provide optional arguments
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo ${2-} $1
}

cd $GOPATH/src/github.com/google/trillian

# Get dependencies
printf "Getting dependencies\n"
go get -d ./...

# Copy over the SQL initialization
cp storage/mysql/storage.sql $COMPOSE_DIR/db/01-init.sql
# Prepend a directive to use the appropriate database
sed -i '1iUSE trillian;' $COMPOSE_DIR/db/01-init.sql

cd $GOPATH/src/github.com/google/trillian/server

printf "Building trillian\n"
cd $COMPOSE_DIR
compile github.com/google/trillian/server/trillian_log_server
compile github.com/google/trillian/server/trillian_log_signer
# Rename the compiled CT server to not conflict with the directory name
compile github.com/google/trillian/examples/ct/ct_server "-o ct_server_main"
compile github.com/google/trillian/cmd/createtree
mv createtree ct_server/
mv ct_server_main ct_server/main
mv trillian_log_server log_server/main
mv trillian_log_signer log_signer/main

cd $COMPOSE_DIR/trampoline
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o trampoline trampoline.go
cp trampoline $COMPOSE_DIR/ct_server
cp trampoline $COMPOSE_DIR/log_signer
cp trampoline $COMPOSE_DIR/log_server

cd $COMPOSE_DIR/ct_server
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o launcher launcher.go
