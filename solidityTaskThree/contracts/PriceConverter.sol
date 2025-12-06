// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@chainlink/contracts/src/v0.8/interfaces/AggregatorV3Interface.sol";

library PriceConverter {
    function getPrice(AggregatorV3Interface priceFeed) internal view returns (uint256) {
        (, int256 answer, , , ) = priceFeed.latestRoundData();
        // priceFeed的decimals通常是8，转换为18位小数以匹配ETH
        return uint256(answer) * 1e10; // 10^18 / 10^8 = 10^10
    }

    function getConversionRate(
        uint256 amount,
        AggregatorV3Interface priceFeed
    ) internal view returns (uint256) {
        uint256 price = getPrice(priceFeed);
        return (amount * price) / 1e18;
    }

    function getUSDValue(
        uint256 amount,
        address token,
        mapping(address => address) storage priceFeeds,
        address ethPriceFeed
    ) internal view returns (uint256) {
        address priceFeedAddress;
        
        if (token == address(0)) {
            priceFeedAddress = ethPriceFeed;
        } else {
            priceFeedAddress = priceFeeds[token];
            require(priceFeedAddress != address(0), "Price feed not set");
        }
        
        AggregatorV3Interface priceFeed = AggregatorV3Interface(priceFeedAddress);
        return getConversionRate(amount, priceFeed);
    }
}