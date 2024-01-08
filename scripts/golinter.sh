#!/bin/bash

cd cronjob-trigger && golangci-lint run ./... && goimports -w  . && gofmt -s -w -e -d .

cd ../github-webhook-trigger && golangci-lint run ./... && goimports -w  . && gofmt -s -w -e -d .

cd ../gitlab-webhook-trigger && golangci-lint run ./... && goimports -w  . && gofmt -s -w -e -d .

cd ../grpc-trigger && golangci-lint run ./... && goimports -w  . && gofmt -s -w -e -d .

cd ../kafka-trigger && golangci-lint run ./... && goimports -w  . && gofmt -s -w -e -d .

cd ../process-trigger && golangci-lint run ./... && goimports -w  . && gofmt -s -w -e -d .

cd ../rest-trigger && golangci-lint run ./... && goimports -w  . && gofmt -s -w -e -d .
