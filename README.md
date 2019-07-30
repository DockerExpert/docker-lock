# Docker Lock

# About
Create lockfiles for docker to ensure reproducible builds.

# Requirements
* docker version >= 19.03
* Recommended, nightly build

# Run
* `docker lock generate`
* `docker lock verify`

# Install
* `mkdir -p ~/.docker/cli-plugins`
* `go get`
* `go build -o ~/.docker/cli-plugins/docker-lock`
