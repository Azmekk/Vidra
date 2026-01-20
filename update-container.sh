#!/bin/bash

# Update the repository code
git fetch && git pull

# Pull the latest pre-built images
docker compose pull

# Restart the containers
docker compose up -d
