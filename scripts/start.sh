# scripts/start.sh
#!/bin/bash

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║     SpiderSpeed Nexus L2 - Full Stack with Real P4 + ZK     ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Ждем, пока Anvil будет готов
echo " Waiting for Ethereum (Anvil) to be ready..."
while ! curl -s http://anvil:8545 > /dev/null; do
    sleep 1
done
echo " Ethereum (Anvil) is ready"

# Создаем виртуальные интерфейсы для P4
echo "🔧 Setting up virtual interfaces for P4..."
ip link add veth0 type veth peer name veth1 2>/dev/null || true
ip link add veth2 type veth peer name veth3 2>/dev/null || true
ip link set veth0 up
ip link set veth2 up

# Запускаем валидатор
echo " Starting Validator..."
/app/bin/validator -port 8001 &
VALIDATOR_PID=$!

# Запускаем мультиплексор
echo " Starting Multiplexer..."
/app/bin/multiplexer -port 8002 -validator localhost:8001 &
MUX_PID=$!

echo ""
echo " ALL SERVICES RUNNING!"
echo ""
echo " Service endpoints:"
echo "   Ethereum (Anvil):   http://localhost:8545"
echo "   P4 Switch (bmv2):   localhost:9090 (Thrift)"
echo "   SP1 Prover:         localhost:8080"
echo "   Validator:          localhost:8001"
echo "   Multiplexer:        localhost:8002"
echo ""
echo " To test inside container:"
echo "   docker exec -it nexus-app /app/bin/client -generate"
echo "   docker exec -it nexus-app /app/bin/client -from <key> -to <key> -amount 10 -count 5"
echo ""
echo "Press Ctrl+C to stop all services"

wait $VALIDATOR_PID $MUX_PID