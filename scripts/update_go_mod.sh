#!/bin/bash

echo "Tidy github webhook trigger go.mod..."
cd github-webhook-trigger
go mod tidy

echo "Done"