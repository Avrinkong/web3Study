// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./AuctionV1.sol";

contract AuctionV2 is AuctionV1 {
    struct DynamicFeeTier {
        uint256 minAmountUSD; // 最小金额（美元，18位小数）
        uint256 feePercentage; // 手续费百分比，100 = 1%
    }
    
    DynamicFeeTier[] public feeTiers;
    
    event FeeTierAdded(uint256 minAmountUSD, uint256 feePercentage);
    event FeeTierRemoved(uint256 index);
    
    function initializeV2() public reinitializer(2) {
        // 初始化动态手续费层级
        feeTiers.push(DynamicFeeTier(0, 250)); // 0-$1000: 2.5%
        feeTiers.push(DynamicFeeTier(1000 * 1e18, 200)); // $1000-$10000: 2.0%
        feeTiers.push(DynamicFeeTier(10000 * 1e18, 150)); // $10000-$50000: 1.5%
        feeTiers.push(DynamicFeeTier(50000 * 1e18, 100)); // $50000+: 1.0%
    }
    
    function addFeeTier(uint256 minAmountUSD, uint256 feePercentage) external onlyOwner {
        require(feePercentage <= 1000, "Fee too high");
        
        feeTiers.push(DynamicFeeTier(minAmountUSD, feePercentage));
        _sortFeeTiers();
        
        emit FeeTierAdded(minAmountUSD, feePercentage);
    }
    
    function removeFeeTier(uint256 index) external onlyOwner {
        require(index < feeTiers.length, "Invalid index");
        
        for (uint256 i = index; i < feeTiers.length - 1; i++) {
            feeTiers[i] = feeTiers[i + 1];
        }
        feeTiers.pop();
        
        emit FeeTierRemoved(index);
    }
    
    function getDynamicFee(uint256 usdAmount) public view returns (uint256) {
        uint256 applicableFee = feePercentage;
        
        for (int256 i = int256(feeTiers.length) - 1; i >= 0; i--) {
            if (usdAmount >= feeTiers[uint256(i)].minAmountUSD) {
                applicableFee = feeTiers[uint256(i)].feePercentage;
                break;
            }
        }
        
        return applicableFee;
    }
    
    function endAuction(uint256 auctionId) public override {
        Auction storage auction = auctions[auctionId];
        
        require(auction.status == AuctionStatus.ACTIVE, "Auction not active");
        require(
            block.timestamp >= auction.endTime || msg.sender == auction.seller,
            "Cannot end early"
        );
        
        auction.status = AuctionStatus.ENDED;
        
        if (auction.highestBidder != address(0)) {
            IERC721 nft = IERC721(auction.nftAddress);
            nft.transferFrom(address(this), auction.highestBidder, auction.tokenId);
            
            // 计算美元价值并获取动态手续费
            uint256 usdAmount = getBidAmountInUSD(auction.paymentToken, auction.highestBid);
            uint256 dynamicFeePercentage = getDynamicFee(usdAmount);
            uint256 fee = auction.highestBid * dynamicFeePercentage / 10000;
            uint256 sellerAmount = auction.highestBid - fee;
            
            if (auction.paymentToken == address(0)) {
                if (fee > 0) {
                    payable(feeRecipient).transfer(fee);
                }
                payable(auction.seller).transfer(sellerAmount);
            } else {
                IERC20 token = IERC20(auction.paymentToken);
                if (fee > 0) {
                    token.safeTransfer(feeRecipient, fee);
                }
                token.safeTransfer(auction.seller, sellerAmount);
            }
            
            emit AuctionEnded(auctionId, auction.highestBidder, auction.highestBid, auction.seller);
        } else {
            IERC721 nft = IERC721(auction.nftAddress);
            nft.transferFrom(address(this), auction.seller, auction.tokenId);
        }
        
        delete nftToAuctionId[auction.nftAddress][auction.tokenId];
    }
    
    function getAllFeeTiers() external view returns (DynamicFeeTier[] memory) {
        return feeTiers;
    }
    
    function _sortFeeTiers() internal {
        for (uint256 i = 0; i < feeTiers.length; i++) {
            for (uint256 j = i + 1; j < feeTiers.length; j++) {
                if (feeTiers[i].minAmountUSD > feeTiers[j].minAmountUSD) {
                    DynamicFeeTier memory temp = feeTiers[i];
                    feeTiers[i] = feeTiers[j];
                    feeTiers[j] = temp;
                }
            }
        }
    }
    
    function version() external pure returns (string memory) {
        return "V2.0.0";
    }
}