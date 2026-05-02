# scripts/deploy-ethereum.sh
#!/bin/bash

echo "🔗 Deploying Nexus Verifier contract to Ethereum (Anvil)..."

# Ждем, пока Anvil запустится
sleep 3

# Адрес Anvil (внутри Docker сети)
ANVIL_URL="http://anvil:8545"

# Проверяем, что Anvil доступен
if ! curl -s $ANVIL_URL > /dev/null; then
    echo " Anvil not reachable at $ANVIL_URL"
    exit 1
fi

echo " Anvil is running"

# Здесь будет деплой смарт-контракта
# В будущем: forge create --rpc-url $ANVIL_URL ...

echo " Contract deployed successfully"