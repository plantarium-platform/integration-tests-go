#!/bin/bash

# Define the HAProxy runtime directory
HAPROXY_RUN_DIR="testdata/temp/.haproxy-run"

# Create the HAProxy runtime directory if it doesn't exist
if [ ! -d "$HAPROXY_RUN_DIR" ]; then
    mkdir -p "$HAPROXY_RUN_DIR"
fi

# Copy the contents of the resources/haproxy directory to the runtime directory
cp -r ./resources/haproxy/* "$HAPROXY_RUN_DIR"

# Run the Docker container with HAProxy
docker run -d --name my-haproxy \
    --network=host \
    -v "${PWD}/$HAPROXY_RUN_DIR:/usr/local/etc/haproxy:rw" \
    haproxytech/haproxy-ubuntu
