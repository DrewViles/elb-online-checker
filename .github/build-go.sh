#!/bin/bash
set -euxo pipefail

# renovate: datasource=github-releases depName=norwoodj/helm-docs
GO_VERSION=1.5.0

# install helm-docs
wget https://golang.org/dl/go1.15.8.linux-amd64.tar.gz -O /tmp/go.tar.gz
tar -C /tmp/ -xzf /tmp/go.tar.gz
export PATH=$PATH:/tmp/go/bin
go version

# validate docs
go build -o elb-online-checker /tmp/elb-online-checker
