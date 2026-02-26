# Fork + Merge test (PositionsMerge subscription)

Test **SubscribeFilterLogs** (realtime, same as prediction-exchange) against a chain that has a PositionsMerge event.

## Run test with real data (Predict BNB RPC)

Using a known block that contains a PositionsMerge log on BSC (Predict ConditionalTokens):

```bash
cd "$(go list -m -f '{{.Dir}}' github.com/ivanzzeth/ethclient)/.."
export SUBSCRIBE_MERGE_TEST_RPC="YOUR_RPC"
export CT_ADDRESS=0x22DA1810B194ca018378464a58f6Ac2B10C9d244
export STAKEHOLDER_ADDRESS=0x2d4370431b5cd3d2ee4ab11f66369b8ed424093e
export FROM_BLOCK=83250010   # first scan range includes block 83250011 (Merge tx)
go test -v -count=1 -run TestSubscribeMergeRealtime ./tests/subscriber/
```

Test builds the same FilterQuery as Predict's **WatchPositionsMerge**, calls **SubscribeFilterLogs** (realtime), and asserts at least one PositionsMerge log is received. Skip if `SUBSCRIBE_MERGE_TEST_RPC` is unset.

## Optional: fork and fetch tx logs

```bash
export RPC="${PREDICT_RPC_URL}"
export TXHASH=0xdfe49b6b3051870c1db60f5647c73ad444dfcdd4fd2e0e86fa989409d9b033ce
./fork_and_fetch_logs.sh
```
