// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./AuctionV1.sol";

contract AuctionV2 is AuctionV1 {
    // 动态手续费结构
    struct DynamicFeeTier {
        uint256 minAmount; // 最小金额（以USD计）
        uint256 feePercentage; // 手续费百分比，100 = 1%
    }
    
    DynamicFeeTier[] public feeTiers;
    
    // 事件
    event FeeTierAdded(uint256 minAmount, uint256 feePercentage);
    event FeeTierRemoved(uint256 index);
    
    // 初始化函数（用于升级后初始化新变量）
    function initializeV2() public reinitializer(2) {
        // 添加默认手续费层级
        feeTiers.push(DynamicFeeTier({
            minAmount: 0,
            feePercentage: 250 // 2.5%
        }));
        feeTiers.push(DynamicFeeTier({
            minAmount: 1000 * 1e18, // $1000
            feePercentage: 200 // 2.0%
        }));
        feeTiers.push(DynamicFeeTier({
            minAmount: 10000 * 1e18, // $10000
            feePercentage: 150 // 1.5%
        }));
        feeTiers.push(DynamicFeeTier({
            minAmount: 100000 * 1e18, // $100000
            feePercentage: 100 // 1.0%
        }));
    }
    
    // 添加手续费层级
    function addFeeTier(uint256 minAmount, uint256 feePercentage) external onlyOwner {
        require(feePercentage <= 1000, "Fee too high");
        
        feeTiers.push(DynamicFeeTier({
            minAmount: minAmount,
            feePercentage: feePercentage
        }));
        
        // 按金额排序
        _sortFeeTiers();
        
        emit FeeTierAdded(minAmount, feePercentage);
    }
    
    // 移除手续费层级
    function removeFeeTier(uint256 index) external onlyOwner {
        require(index < feeTiers.length, "Invalid index");
        
        for (uint256 i = index; i < feeTiers.length - 1; i++) {
            feeTiers[i] = feeTiers[i + 1];
        }
        feeTiers.pop();
        
        emit FeeTierRemoved(index);
    }
    
    // 获取动态手续费
    function getDynamicFee(uint256 usdAmount) public view returns (uint256) {
        uint256 applicableFee = feePercentage; // 默认手续费
        
        for (int256 i = int256(feeTiers.length) - 1; i >= 0; i--) {
            if (usdAmount >= feeTiers[uint256(i)].minAmount) {
                applicableFee = feeTiers[uint256(i)].feePercentage;
                break;
            }
        }
        
        return applicableFee;
    }
    
    // 重写结束拍卖函数以使用动态手续费
    function endAuction(uint256 auctionId) external override {
        Auction storage auction = auctions[auctionId];
        
        require(auction.status == AuctionStatus.ACTIVE, "Auction not active");
        require(block.timestamp >= auction.endTime || msg.sender == auction.seller, "Cannot end early");
        
        auction.status = AuctionStatus.ENDED;
        
        if (auction.highestBidder != address(0)) {
            // 转移NFT给获胜者
            IERC721 nft = IERC721(auction.nftAddress);
            nft.transferFrom(address(this), auction.highestBidder, auction.tokenId);
            
            // 计算美元价值
            uint256 usdAmount = getBidAmountInUSD(auction.paymentToken, auction.highestBid);
            
            // 使用动态手续费
            uint256 dynamicFeePercentage = getDynamicFee(usdAmount);
            uint256 fee = auction.highestBid * dynamicFeePercentage / 10000;
            uint256 sellerAmount = auction.highestBid - fee;
            
            // 转移资金
            if (auction.paymentToken == address(0)) {
                // ETH
                if (fee > 0) {
                    payable(feeRecipient).transfer(fee);
                }
                payable(auction.seller).transfer(sellerAmount);
            } else {
                // ERC20
                IERC20 token = IERC20(auction.paymentToken);
                if (fee > 0) {
                    token.safeTransfer(feeRecipient, fee);
                }
                token.safeTransfer(auction.seller, sellerAmount);
            }
            
            emit AuctionEnded(auctionId, auction.highestBidder, auction.highestBid, auction.seller);
        } else {
            // 无人出价，退还NFT
            IERC721 nft = IERC721(auction.nftAddress);
            nft.transferFrom(address(this), auction.seller, auction.tokenId);
        }
        
        // 清理映射
        delete nftToAuctionId[auction.nftAddress][auction.tokenId];
    }
    
    // 获取所有手续费层级
    function getAllFeeTiers() external view returns (DynamicFeeTier[] memory) {
        return feeTiers;
    }
    
    // 内部函数：排序手续费层级
    function _sortFeeTiers() internal {
        for (uint256 i = 0; i < feeTiers.length; i++) {
            for (uint256 j = i + 1; j < feeTiers.length; j++) {
                if (feeTiers[i].minAmount > feeTiers[j].minAmount) {
                    DynamicFeeTier memory temp = feeTiers[i];
                    feeTiers[i] = feeTiers[j];
                    feeTiers[j] = temp;
                }
            }
        }
    }
    
    // 版本标识
    function version() external pure returns (string memory) {
        return "V2.0.0";
    }
}