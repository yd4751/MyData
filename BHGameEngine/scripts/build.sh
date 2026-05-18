#!/bin/bash

echo "Building OpenWorld Server..."

mkdir -p bin

echo "Building Data Service..."
go build -o bin/dataservice ./cmd/dataservice

echo "Building Login Server..."
go build -o bin/login ./cmd/login

echo "Building Gate Server..."
go build -o bin/gate ./cmd/gate

echo "Building GridMap Server..."
go build -o bin/gridmap ./cmd/gridmap

echo "Building Logic Server..."
go build -o bin/logic ./cmd/logic

echo "Building Battle Server..."
go build -o bin/battle ./cmd/battle

echo "Building Cross Server..."
go build -o bin/cross ./cmd/cross

echo "Building GM Server..."
go build -o bin/gm ./cmd/gm

echo "Building Web Server..."
go build -o bin/webserver ./cmd/webserver

echo "All servers built successfully!"