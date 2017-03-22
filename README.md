# Astra

For Cornell CS 5152

`trillian/`: Docker & docker-compose configuration for Trillian,
Google's certificate transparency server (written in Go).

Usage:

    % ./update_trillian.sh
    % docker-compose up

`submit/`: example of submitting certificates from a cfssl `certdb` to
Trillian (or another CT server).
