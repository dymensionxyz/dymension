// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.19;
import "forge-std/Test.sol";
import "../src/MyContract.sol";

contract MyContractTest is Test {
    MyContract private mc;

    function setUp() public {
        mc = new MyContract();
    }

    function testInitialValue() public {
        assertEq(mc.value(), 0);
    }

    function testSetValue() public {
        mc.setValue(123);
        assertEq(mc.value(), 123);
    }
}
