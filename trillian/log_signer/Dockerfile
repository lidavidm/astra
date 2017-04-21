FROM scratch
ADD main /
ADD trampoline /
VOLUME ["/config"]
# Use ENTRYPOINT so that the user can pass in additional arguments to the server
ENTRYPOINT ["/trampoline", "--target", "/main", "--config", "/config/signer_params.json"]
CMD []