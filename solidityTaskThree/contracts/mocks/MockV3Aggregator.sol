// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@chainlink/contracts/src/v0.8/interfaces/AggregatorV3Interface.sol";

contract MockV3Aggregator is AggregatorV3Interface {
    uint8 private _decimals;
    int256 private _latestAnswer;
    
    constructor(uint8 decimals_, int256 latestAnswer_) {
        _decimals = decimals_;
        _latestAnswer = latestAnswer_;
    }
    
    function decimals() external view override returns (uint8) {
        return _decimals;
    }
    
    function latestRoundData() external view override returns (
        uint80 roundId,
        int256 answer,
        uint256 startedAt,
        uint256 updatedAt,
        uint80 answeredInRound
    ) {
        return (0, _latestAnswer, block.timestamp, block.timestamp, 0);
    }
    
    // 以下函数仅用于实现接口
    function description() external pure override returns (string memory) {
        return "Mock Price Feed";
    }
    
    function version() external pure override returns (uint256) {
        return 1;
    }
    
    function getRoundData(uint80) external pure override returns (
        uint80 roundId,
        int256 answer,
        uint256 startedAt,
        uint256 updatedAt,
        uint80 answeredInRound
    ) {
        revert("Not implemented");
    }
}