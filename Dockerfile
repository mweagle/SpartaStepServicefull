FROM alpine:3.8
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

ARG SPARTA_DOCKER_BINARY
ADD $SPARTA_DOCKER_BINARY /SpartaServicefull
CMD ["/SpartaServicefull", "fargateTask"]