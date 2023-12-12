#!/bin/bash

echo "Tidy github webhook trigger go.mod..."
cd github-webhook-trigger
go mod tidy

cd ..
echo "Tidy gitlab webhook trigger go.mod..."
cd gitlab-webhook-trigger
go mod tidy

cd ..
echo "Tidy cronjob trigger go.mod..."
cd cronjob-trigger
go mod tidy

cd ..
echo "Tidy grpc trigger go.mod..."
cd grpc-trigger
go mod tidy

cd ..
echo "Tidy kafka trigger go.mod..."
cd kafka-trigger
go mod tidy

cd ..
echo "Tidy process trigger go.mod..."
cd process-trigger
go mod tidy

cd ..
echo "Tidy rest trigger go.mod..."
cd rest-trigger
go mod tidy

echo "Done"
