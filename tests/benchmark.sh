#!/bin/bash
# Zero-Downtime Benchmarking Orchestrator

# Setup colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}[*] Checking for wrk utility...${NC}"
if ! command -v wrk &> /dev/null; then
    echo -e "${GREEN}[*] Installing wrk load testing tool...${NC}"
    sudo apt-get update && sudo apt-get install -y wrk
fi

# Ensure at least one environment is active first (bootstrap if needed)
echo -e "${GREEN}[*] Ensuring baseline active deployment exists...${NC}"
ansible-playbook ../ansible/deploy.yml --tags "common,app_deploy,health_check,traffic_switch"

echo -e "${GREEN}[*] Starting background load testing (Duration: 20 seconds, Concurrency: 50)...${NC}"
# Run wrk targeting Nginx gateway on port 80
wrk -t4 -c50 -d20s -s wrk_downtime.lua http://127.0.0.1/health > wrk_output.log &
WRK_PID=$!

# Wait 4 seconds to collect initial metrics under load
sleep 4

echo -e "${GREEN}[*] TRIGGERING BLUE-GREEN DEPLOYMENT MID-BENCHMARK...${NC}"
ansible-playbook ../ansible/deploy.yml

echo -e "${GREEN}[*] Waiting for load testing to conclude...${NC}"
wait $WRK_PID

echo -e "${GREEN}[*] Benchmark complete. Logging results:${NC}"
cat wrk_output.log
rm -f wrk_output.log
