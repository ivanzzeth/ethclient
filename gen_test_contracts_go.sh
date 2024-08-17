solc --abi --bin ./contracts/Test.sol -o ./contracts --overwrite
abigen --bin=./contracts/Test.bin --abi=./contracts/Test.abi --pkg=contracts --out=./contracts/test_contract.go