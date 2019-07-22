# Docker Lock

# Requirements
docker version >= 19.03

# Run
docker lock generate
docker lock verify

# Install
mkdir -p ~/.docker/cli-plugins
docker build -o ~/.docker/cli-plugins/docker-lock
