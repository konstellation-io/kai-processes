#!/bin/bash

protoc -I=. \
  --go_out="./proto" \
  --go_opt=paths=source_relative *.proto \
  --go-grpc_out="./proto" \
  --go-grpc_opt=paths=source_relative *.proto \

echo "Done"