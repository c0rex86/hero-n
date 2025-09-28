#!/bin/bash

cd server

echo "building server..."
go build -o heroin-server ./cmd/heroin-server

if [ ! -f heroin-server ]; then
    echo "build failed"
    exit 1
fi

echo "starting ipfs daemon..."
docker run -d --name ipfs-test \
    -p 5001:5001 \
    -p 4001:4001 \
    ipfs/go-ipfs:latest

sleep 5

echo "starting server..."
HEROIN_CONFIG=configs/config.example.yaml ./heroin-server &
SERVER_PID=$!

sleep 3

echo "testing grpc health..."
grpcurl -plaintext localhost:50051 list

echo "server pid: $SERVER_PID"
echo "press enter to stop..."
read

kill $SERVER_PID
docker stop ipfs-test
docker rm ipfs-test

echo "cleanup done"
