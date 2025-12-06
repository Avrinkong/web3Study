// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;
import "./AuctionV1.sol";

contract AuctionV2 is AuctionV1 {
    // 新增状态变量：动态手续费层级
    struct FeeTier {
        uint256 minUSD; // 最小金额（USD，8位小数）
        uint256 feePercent; // 费率 (100 = 1%)
    }
    FeeTier[] public feeTiers;

    // 新增事件
    event FeeTierAdded(uint256 minUSD, uint256 feePercent);

    // 初始化V2的新功能（在升级后调用）
    function initializeV2() public reinitializer(2) {
        // 初始化示例手续费层级
        addFeeTier(0, 250);       // 0-$1000: 2.5%
        addFeeTier(1000 * 1e8, 200); // $1000-$5000: 2.0%
        addFeeTier(5000 * 1e8, 150); // $5000以上: 1.5%
    }

    // 添加手续费层级（仅所有者）
    function addFeeTier(uint256 minUSD, uint256 feePercent) public onlyOwner {
        feeTiers.push(FeeTier(minUSD, feePercent));
        emit FeeTierAdded(minUSD, feePercent);
    }

    // 根据USD金额获取动态费率
    function getDynamicFee(uint256 usdAmount) public view returns (uint256) {
        uint256 applicableFee = feePercentage; // 默认费率
        for (uint256 i = feeTiers.length; i > 0; i--) {
            if (usdAmount >= feeTiers[i-1].minUSD) {
                applicableFee = feeTiers[i-1].feePercent;
                break;
            }
        }
        return applicableFee;
    }

    // 重写结束拍卖函数，应用动态手续费
    function endAuction(uint256 auctionId) public override {
        Auction storage auction = auctions[auctionId];
        require(block.timestamp >= auction.endTime, "Auction not ended");
        require(!auction.ended, "Auction already ended");
        require(msg.sender == auction.seller || msg.sender == owner(), "Not authorized");

        auction.ended = true;

        if (auction.highestBidder != address(0)) {
            // 计算当前出价的美元价值
            uint256 usdValue = getBidValueInUSD(auction.paymentToken, auction.highestBid);
            // 根据美元价值获取动态费率
            uint256 dynamicFee = getDynamicFee(usdValue);
            uint256 fee = (auction.highestBid * dynamicFee) / 10000;
            uint256 sellerProceeds = auction.highestBid - fee;

            // 转移资金
            if (auction.paymentToken == address(0)) {
                payable(feeRecipient).transfer(fee);
                payable(auction.seller).transfer(sellerProceeds);
            } else {
                IERC20 token = IERC20(auction.paymentToken);
                token.transfer(feeRecipient, fee);
                token.transfer(auction.seller, sellerProceeds);
            }
            // 转移NFT
            IERC721(auction.nftContract).transferFrom(address(this), auction.highestBidder, auction.tokenId);
            emit AuctionEnded(auctionId, auction.highestBidder, auction.highestBid);
        } else {
            IERC721(auction.nftContract).transferFrom(address(this), auction.seller, auction.tokenId);
        }
    }
}