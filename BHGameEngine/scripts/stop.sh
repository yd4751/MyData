#!/bin/bash

echo "Stopping OpenWorld Server..."

if [ -f pid/dataservice.pid ]; then
    DATA_PID=$(cat pid/dataservice.pid)
    kill $DATA_PID 2>/dev/null
    echo "Data Service stopped"
fi

if [ -f pid/login.pid ]; then
    LOGIN_PID=$(cat pid/login.pid)
    kill $LOGIN_PID 2>/dev/null
    echo "Login Server stopped"
fi

if [ -f pid/gate.pid ]; then
    GATE_PID=$(cat pid/gate.pid)
    kill $GATE_PID 2>/dev/null
    echo "Gate Server stopped"
fi

if [ -f pid/gridmap.pid ]; then
    GRIDMAP_PID=$(cat pid/gridmap.pid)
    kill $GRIDMAP_PID 2>/dev/null
    echo "GridMap Server stopped"
fi

if [ -f pid/logic.pid ]; then
    LOGIC_PID=$(cat pid/logic.pid)
    kill $LOGIC_PID 2>/dev/null
    echo "Logic Server stopped"
fi

if [ -f pid/battle.pid ]; then
    BATTLE_PID=$(cat pid/battle.pid)
    kill $BATTLE_PID 2>/dev/null
    echo "Battle Server stopped"
fi

if [ -f pid/cross.pid ]; then
    CROSS_PID=$(cat pid/cross.pid)
    kill $CROSS_PID 2>/dev/null
    echo "Cross Server stopped"
fi

if [ -f pid/gm.pid ]; then
    GM_PID=$(cat pid/gm.pid)
    kill $GM_PID 2>/dev/null
    echo "GM Server stopped"
fi

if [ -f pid/web.pid ]; then
    WEB_PID=$(cat pid/web.pid)
    kill $WEB_PID 2>/dev/null
    echo "WEB Server stopped"
fi

rm -rf pid/*.pid

echo "All servers stopped successfully!"