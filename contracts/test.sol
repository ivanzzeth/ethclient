// SPDX-License-Identifier: UNLICENSED
pragma solidity >=0.8.0;

contract Test {
    uint256 public counter;

    event CounterUpdated(uint256 counter);
    event FuncEvent1(string arg1, uint256 arg2, bytes arg3);

    function testFunc1(string memory arg1, uint256 arg2, bytes memory arg3) public {
        emit FuncEvent1(arg1, arg2, arg3);
        counter += 1;
        emit CounterUpdated(counter);
    }

    function testReverted(bool r) public pure {
        if (r) {
            revert("test reverted");
        }
    }

    function testControlledReverted() public view {
        if (block.number % 4 == 0) {
            revert("test reverted");
        }
    }
}