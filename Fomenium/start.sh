#!/bin/bash
set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}  Fomenium - Hardware-Accelerated L2${NC}"
echo -e "${BLUE}============================================${NC}"
echo ""

command -v docker >/dev/null 2>&1 || { echo -e "${RED}Docker required${NC}"; exit 1; }

echo -e "${YELLOW}Building and starting...${NC}"
docker compose down -v 2>/dev/null || true
docker compose build 2>&1 | tail -3
docker compose up -d

echo -e "${YELLOW}Waiting for services (60s max)...${NC}"
for i in $(seq 1 20); do
    if curl -s http://localhost:8080/health 2>/dev/null | grep -q "ok"; then
        echo -e "\n${GREEN}Ready!${NC}"
        break
    fi
    echo -n "."
    sleep 3
done

# Прогреваем P4
echo ""
echo -e "${YELLOW}Warming up P4...${NC}"
curl -s -X POST http://localhost:8080/aggregate \
    -H 'Content-Type: application/json' \
    -d '{"batch_id":0,"payment_id":0,"from_account":1,"to_account":2,"amount":1,"hash":"warmup"}' > /dev/null 2>&1
sleep 1

echo ""
echo -e "${BLUE}--- Status ---${NC}"
HEALTH=$(curl -s http://localhost:8080/health)
P4=$(echo "$HEALTH" | jq -r '.p4_real')
ETH=$(echo "$HEALTH" | jq -r '.eth_connected')
VAL=$(echo "$HEALTH" | jq -r '.validator_ok')

echo -e "P4 Switch:    ${GREEN}$([ "$P4" = "true" ] && echo "bmv2" || echo "simulation")${NC}"
echo -e "ZK Prover:    ${GREEN}SP1 DevNet${NC}"
echo -e "Validator:    ${GREEN}$([ "$VAL" = "true" ] && echo "running" || echo "down")${NC}"
echo -e "Multiplexer:  ${GREEN}$([ "$ETH" = "true" ] && echo "connected" || echo "down")${NC}"

echo ""
echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}  BENCHMARK: 5 Transactions${NC}"
echo -e "${BLUE}============================================${NC}"
echo ""

RESULT=$(curl -s -X POST http://localhost:8080/api/benchmark \
    -H 'Content-Type: application/json' \
    -d '{"txCount":5,"amount":100}')

NEXUS_MS=$(echo "$RESULT" | jq -r '.nexus.duration_ms')
ANVIL_MS=$(echo "$RESULT" | jq -r '.anvil.duration_ms')
SPEEDUP=$(echo "$RESULT" | jq -r '.speedup')
CONFIRMED=$(echo "$RESULT" | jq -r '.anvil.confirmed')
ZK_MODE=$(echo "$RESULT" | jq -r '.zk_mode')

echo -e "${GREEN}Fomenium:${NC}     ${NEXUS_MS}ms (1 batched tx)"
echo -e "${YELLOW}Ethereum L1:${NC}  ${ANVIL_MS}ms (5 separate, ${CONFIRMED}/5 confirmed)"
echo -e "${BLUE}Speedup:${NC}      $(printf "%.1f" $SPEEDUP)x | ZK: ${ZK_MODE}"

echo ""
echo -e "${BLUE}============================================${NC}"
echo -e "  curl http://localhost:8080/health | jq ."
echo -e "  curl -X POST http://localhost:8080/api/benchmark \\"
echo -e "    -H 'Content-Type: application/json' \\"
echo -e "    -d '{\"txCount\":50,\"amount\":10}' | jq ."
echo -e "${BLUE}============================================${NC}"
