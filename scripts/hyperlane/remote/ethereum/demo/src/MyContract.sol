// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.19;

contract MyContract {
    uint256 public value;

    function setValue(uint256 _value) public {
        value = _value;
    }
}