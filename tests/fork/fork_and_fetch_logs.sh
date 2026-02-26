#!/usr/bin/env bash
# Fork chain and fetch logs of a tx that triggered PositionsMerge.
# Usage:
#   RPC=https://opbnb-mainnet-rpc.bnbchain.org TXHASH=0x... [BLOCK=123] ./fork_and_fetch_logs.sh
# Then run anvil in background (or in another terminal) and this script will fetch the receipt logs.

set -e
RPC="${RPC:-}"
TXHASH="${TXHASH:-}"
BLOCK="${BLOCK:-}"
ANVIL_PORT="${ANVIL_PORT:-8545}"
RPC_URL="http://127.0.0.1:${ANVIL_PORT}"

if [ -z "$RPC" ]; then
  echo "Set RPC (e.g. https://opbnb-mainnet-rpc.bnbchain.org)"
  exit 1
fi

echo "Starting anvil fork: $RPC ${BLOCK:+--fork-block-number $BLOCK}"
anvil --fork-url "$RPC" ${BLOCK:+--fork-block-number $BLOCK} --port "$ANVIL_PORT" &
ANVIL_PID=$!
trap "kill $ANVIL_PID 2>/dev/null || true" EXIT

# Wait for anvil
for i in 1 2 3 4 5 6 7 8 9 10; do
  if cast block-number --rpc-url "$RPC_URL" 2>/dev/null; then break; fi
  sleep 1
done
cast block-number --rpc-url "$RPC_URL" >/dev/null || { echo "anvil not ready"; exit 1; }

if [ -n "$TXHASH" ]; then
  echo "Fetching receipt and logs for $TXHASH"
  cast receipt "$TXHASH" --rpc-url "$RPC_URL" -j
  echo ""
  echo "Logs only:"
  cast receipt "$TXHASH" --rpc-url "$RPC_URL" -j | jq '.logs'
else
  echo "TXHASH not set; anvil is running. Set TXHASH and re-run to fetch logs, or trigger Merge manually then run Go test."
  wait $ANVIL_PID
fi
