// SPDX-License-Identifier: UNLICENSED
pragma solidity >=0.8.0;

contract Test {
    uint256 public counter;

    error TestRevert(uint a, uint b);

    event CounterUpdated(uint256 counter);
    event FuncEvent1(string arg1, uint256 arg2, bytes arg3);

    function testFunc1(string memory arg1, uint256 arg2, bytes memory arg3) public {
        emit FuncEvent1(arg1, arg2, arg3);
        counter += 1;
        emit CounterUpdated(counter);
    }

    function testReverted(bool r) public pure {
        if (r) {
            revert TestRevert(1, 2);
        }
    }

    function testRandomlyReverted() public view {
        if (block.number % 4 == 0) {
            revert TestRevert(1, 2);
        }
    }
}