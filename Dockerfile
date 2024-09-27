# This Dockerfile might not follow best practices, but that is an intentional
# choice to have this Dockerfile use the install scripts that users use in the tutorial.

ARG ROLLKIT_VERSION="v0.13.7"

FROM ubuntu:latest

# Install system dependencies
RUN apt update && apt install -y bash curl jq git make sed ranger vim golang && apt clean

# Set the working directory
WORKDIR /app

# Make sure GOPATH is set
ENV GOPATH /usr/local/go
ENV PATH $GOPATH/bin:$PATH

# Install Rollkit dependencies
RUN (curl -sSL https://rollkit.dev/install.sh | sh -s ${ROLLKIT_VERSION}) && go clean -modcache

# Install Artela rollup
RUN mkdir -p /app/artela-rollkit
COPY . /app/artela-rollkit
COPY ./lazy_config /root/.artroll

# Update the working directory
WORKDIR /app/artela-rollkit

# Initialize the Rollkit configuration
RUN rollkit toml init

# Edit rollkit.toml config_dir
RUN sed -i 's/config_dir = "artroll"/config_dir = "\/root\/\.artroll"/g' rollkit.toml

# Run base rollkit command to download packages
RUN rollkit && go clean -modcache

# Keep the container running
CMD tail -F /dev/null
