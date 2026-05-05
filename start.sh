#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}  Fomenium - Hardware-Accelerated L2${NC}"
echo -e "${BLUE}============================================${NC}"
echo ""

# Check prerequisites
command -v docker >/dev/null 2>&1 || { echo -e "${RED}Docker required. Install: https://docker.com${NC}"; exit 1; }
command -v jq >/dev/null 2>&1 || { echo -e "${YELLOW}Installing jq...${NC}"; sudo apt install -y jq 2>/dev/null || true; }

# Cleanup old containers
echo -e "${YELLOW}Cleaning up...${NC}"
docker compose down -v 2>/dev/null || true

# Build all services
echo -e "${YELLOW}Building all services...${NC}"
docker compose build --no-cache 2>&1 | tail -5

# Start all services
echo -e "${YELLOW}Starting services...${NC}"
docker compose up -d

# Wait for healthy status
echo -e "${YELLOW}Waiting for initialization...${NC}"
WAIT_TIME=0
MAX_WAIT=120
while [ $WAIT_TIME -lt $MAX_WAIT ]; do
    HEALTHY=$(docker compose ps 2>/dev/null | grep -c "healthy" || echo "0")
    RUNNING=$(docker compose ps 2>/dev/null | grep -c "Up" || echo "0")
    
    if [ "$HEALTHY" -ge 3 ] && [ "$RUNNING" -ge 5 ]; then
        echo -e "\n${GREEN}All services ready!${NC}"
        break
    fi
    
    echo -n "."
    sleep 3
    WAIT_TIME=$((WAIT_TIME + 3))
done

echo ""
echo ""

# ============================================
# Component Status
# ============================================
echo -e "${BLUE}--- Component Status ---${NC}"

# Anvil
ANVIL_BLOCK=$(curl -s -X POST http://localhost:8545 \
    -H 'Content-Type: application/json' \
    -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result' 2>/dev/null)
if [ "$ANVIL_BLOCK" != "null" ] && [ -n "$ANVIL_BLOCK" ]; then
    echo -e "${GREEN}Anvil (Ethereum L1):${NC}    Block $ANVIL_BLOCK"
else
    echo -e "${RED}Anvil:${NC}                  NOT READY"
fi

# P4
if docker exec nexus-p4 sh -c 'echo > /dev/tcp/localhost/9090' 2>/dev/null; then
    echo -e "${GREEN}P4 Switch (bmv2):${NC}      Running"
else
    echo -e "${RED}P4 Switch:${NC}             NOT READY"
fi

# SP1
SP1_STATUS=$(curl -s http://localhost:8081/health 2>/dev/null | jq -r '.mode' 2>/dev/null)
if [ -n "$SP1_STATUS" ]; then
    echo -e "${GREEN}ZK Prover (SP1 DevNet):${NC} $SP1_STATUS"
else
    echo -e "${RED}ZK Prover:${NC}             NOT READY"
fi

# Validator
if docker exec nexus-validator sh -c 'echo > /dev/tcp/localhost/8001' 2>/dev/null; then
    echo -e "${GREEN}Validator:${NC}             Running"
else
    echo -e "${RED}Validator:${NC}             NOT READY"
fi

# Multiplexer
MUX_HEALTH=$(curl -s http://localhost:8080/health 2>/dev/null)
P4_REAL=$(echo "$MUX_HEALTH" | jq -r '.p4_real' 2>/dev/null)
ETH_OK=$(echo "$MUX_HEALTH" | jq -r '.eth_connected' 2>/dev/null)
if [ "$ETH_OK" = "true" ]; then
    echo -e "${GREEN}Multiplexer:${NC}           P4=$P4_REAL ETH=$ETH_OK"
else
    echo -e "${RED}Multiplexer:${NC}           NOT READY"
fi

echo ""

# ============================================
# Benchmark
# ============================================
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
P4_REAL=$(echo "$RESULT" | jq -r '.p4_real')
ZK_MODE=$(echo "$RESULT" | jq -r '.zk_mode')
NEXUS_CONFIRMED=$(echo "$RESULT" | jq -r '.nexus.confirmed')

echo -e "${GREEN}Fomenium L2:${NC}    ${NEXUS_MS}ms | 1 batched tx | confirmed=$NEXUS_CONFIRMED"
echo -e "${YELLOW}Ethereum L1:${NC}   ${ANVIL_MS}ms | 5 separate txs | confirmed=${CONFIRMED}/5"
echo -e "${BLUE}Components:${NC}     P4=$P4_REAL | ZK=$ZK_MODE"
echo ""
echo -e "${BLUE}Speedup: ${SPEEDUP}x faster${NC}"

# Show proof on-chain
echo ""
LATEST_BLOCK=$(curl -s -X POST http://localhost:8545 \
    -H 'Content-Type: application/json' \
    -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo -e "${GREEN}Proof:${NC} Real transactions in Anvil block $LATEST_BLOCK"

echo ""
echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}  Services Running${NC}"
echo -e "${BLUE}============================================${NC}"
echo ""
echo "  Ethereum L1:  http://localhost:8545 (Anvil)"
echo "  P4 Switch:    http://localhost:9090 (bmv2)"
echo "  ZK Prover:    http://localhost:8081 (SP1 DevNet)"
echo "  Validator:    http://localhost:8001"
echo "  Multiplexer:  http://localhost:8080"
echo ""
echo "  Health:  curl http://localhost:8080/health | jq ."
echo "  Bench:   curl -X POST http://localhost:8080/api/benchmark -H 'Content-Type: application/json' -d '{\"txCount\":5,\"amount\":100}' | jq ."
echo -e "${BLUE}============================================${NC}"
