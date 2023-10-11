#!/bin/sh

CGO_ENABLED=0 go build -o build/process main.go
minikube image build -t kai-processes/gitlab-webhook-trigger . -f develop.Dockerfile -p kai-local
