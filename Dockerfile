FROM alpine:3.8
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

# Sparta provides the SPARTA_DOCKER_BINARY argument to the builder
# in order to embed the binary.
# Ref: https://docs.docker.com/engine/reference/builder/
ARG SPARTA_DOCKER_BINARY
ADD $SPARTA_DOCKER_BINARY /SpartaServicefull
CMD ["/SpartaServicefull", "fargateTask"]