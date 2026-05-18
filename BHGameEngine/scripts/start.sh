#!/bin/bash

echo "Starting OpenWorld Server..."

mkdir -p logs
mkdir -p pid

echo "Starting Data Service..."
nohup ./bin/dataservice -config config/config.toml > logs/dataservice.log 2>&1 &
DATA_PID=$!
echo "Data Service started with PID: $DATA_PID"

sleep 2

echo "Starting Login Server..."
nohup ./bin/login -config config/config.toml > logs/login.log 2>&1 &
LOGIN_PID=$!
echo "Login Server started with PID: $LOGIN_PID"

sleep 2

echo "Starting Gate Server..."
nohup ./bin/gate -config config/config.toml > logs/gate.log 2>&1 &
GATE_PID=$!
echo "Gate Server started with PID: $GATE_PID"

sleep 2

echo "Starting GridMap Server..."
nohup ./bin/gridmap -config config/config.toml > logs/gridmap.log 2>&1 &
GRIDMAP_PID=$!
echo "GridMap Server started with PID: $GRIDMAP_PID"

sleep 2

echo "Starting Logic Server..."
nohup ./bin/logic -config config/config.toml > logs/logic.log 2>&1 &
LOGIC_PID=$!
echo "Logic Server started with PID: $LOGIC_PID"

sleep 2

echo "Starting Battle Server..."
nohup ./bin/battle -config config/config.toml > logs/battle.log 2>&1 &
BATTLE_PID=$!
echo "Battle Server started with PID: $BATTLE_PID"

sleep 2

echo "Starting Cross Server..."
nohup ./bin/cross -config config/config.toml > logs/cross.log 2>&1 &
CROSS_PID=$!
echo "Cross Server started with PID: $CROSS_PID"

sleep 2

echo "Starting GM Server..."
nohup ./bin/gm -config config/config.toml > logs/gm.log 2>&1 &
GM_PID=$!
echo "GM Server started with PID: $GM_PID"

sleep 2

echo "Starting WEB Server..."
nohup ./bin/webserver -config config/config.toml > logs/web.log 2>&1 &
WEB_PID=$!
echo "WEB Server started with PID: $WEB_PID"

echo "All servers started successfully!"

echo "$DATA_PID" > pid/dataservice.pid
echo "$LOGIN_PID" > pid/login.pid
echo "$GATE_PID" > pid/gate.pid
echo "$GRIDMAP_PID" > pid/gridmap.pid
echo "$LOGIC_PID" > pid/logic.pid
echo "$BATTLE_PID" > pid/battle.pid
echo "$CROSS_PID" > pid/cross.pid
echo "$GM_PID" > pid/gm.pid
echo "$WEB_PID" > pid/web.pid