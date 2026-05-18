#!/bin/bash

echo "Building OpenWorld Server..."

mkdir -p bin

echo "Building Data Service..."
go build -o bin/dataservice cmd/dataservice/main.go

echo "Building Login Server..."
go build -o bin/login cmd/login/main.go

echo "Building Gate Server..."
go build -o bin/gate cmd/gate/main.go

echo "Building GridMap Server..."
go build -o bin/gridmap cmd/gridmap/main.go

echo "Building Logic Server..."
go build -o bin/logic cmd/logic/main.go

echo "Building Battle Server..."
go build -o bin/battle cmd/battle/main.go

echo "Building Cross Server..."
go build -o bin/cross cmd/cross/main.go

echo "Building GM Server..."
go build -o bin/gm cmd/gm/main.go

echo "Building Web Server..."
go build -o bin/webserver cmd/webserver/main.go

echo "All servers built successfully!"