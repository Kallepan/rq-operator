FROM alpine:latest AS base

# Set Environment Variables
ARG GOLANG_VERSION

# Install Packages
RUN apk add --no-cache \
    make \
    bash \
    curl \
    tar \
    gzip

# Install GoLang
RUN curl -O https://dl.google.com/go/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz

# Create Final docker image
FROM docker:cli

RUN apk add --no-cache \
    make \
    curl \
    bash

COPY --from=base /usr/local/go /usr/local/go

ENV PATH=$PATH:/usr/local/go/bin