# Astra

For Cornell CS 5152

### trillian 

`trillian/`: Docker & docker-compose configuration for Trillian,
Google's certificate transparency server (written in Go).

Usage:

    % ./update_trillian.sh
    % docker-compose up

Configuration:
* ct_config.json
* ct_params.json - Command-line parameters to the CT server personality
* log_params.json - Command-line parameters to the CT log binary
* privkey.pem
* pubkey.pem
* roots.cert - A list of acceptable Root Certificates, see RFC sec 3.1
* signer_params.json - Command-line parameters to the CT log signer service

### submit

`submit/`: example of submitting certificates from a cfssl `certdb` to
Trillian (or another CT server).

### save-ct-roots

`save-ct-roots`: retrieve a running CT server's list of accepted root certificates and save as PEM

Usage:

    % ./save-ct-roots <CT server URL>

### ctmerge

`ctmerge/`

Grab certificates sequentially from one running CT server and submit each to another running CT server.

	Usage: ./ctmerge [options...] <source CT server base URL> <dest CT server base URL>
	  -allow_verification_with_non_compliant_keys
			Allow a SignatureVerifier to use keys which are technically non-compliant with RFC6962.
	  -size int
			Amount of entries to query each iteration (default 64)
	  -start int
			The 0-based leaf entry index to start at

