FROM scratch

VOLUME ["/config"]
ADD main /
ADD trampoline /
ADD launcher /
ADD createtree /

# Use ENTRYPOINT so that the user can pass in additional arguments to the server
ENTRYPOINT ["/launcher"]
CMD []
