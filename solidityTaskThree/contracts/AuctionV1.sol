// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@chainlink/contracts/src/v0.8/interfaces/AggregatorV3Interface.sol";

contract AuctionV1 is Initializable, UUPSUpgradeable, OwnableUpgradeable {
    using SafeERC20 for IERC20;
    
    // 拍卖状态枚举
    enum AuctionStatus { PENDING, ACTIVE, ENDED, CANCELLED }
    
    // 拍卖结构体
    struct Auction {
        uint256 id;
        address seller;
        address nftAddress;
        uint256 tokenId;
        uint256 startTime;
        uint256 endTime;
        uint256 startPrice;
        address paymentToken; // 零地址表示ETH
        address highestBidder;
        uint256 highestBid;
        AuctionStatus status;
    }
    
    // Chainlink价格Feed地址
    mapping(address => address) public priceFeeds; // token => priceFeed
    address public ethUsdPriceFeed;
    
    // 拍卖映射
    mapping(uint256 => Auction) public auctions;
    mapping(address => mapping(uint256 => uint256)) public nftToAuctionId;
    
    // 出价记录
    struct Bid {
        address bidder;
        uint256 amount;
        uint256 timestamp;
        bool refunded;
    }
    
    mapping(uint256 => Bid[]) public auctionBids;
    
    // 手续费
    uint256 public feePercentage; // 百分比，100 = 1%
    address public feeRecipient;
    
    // 事件
    event AuctionCreated(
        uint256 indexed auctionId,
        address indexed seller,
        address indexed nftAddress,
        uint256 tokenId,
        uint256 startTime,
        uint256 endTime,
        uint256 startPrice,
        address paymentToken
    );
    
    event BidPlaced(
        uint256 indexed auctionId,
        address indexed bidder,
        uint256 amount,
        uint256 usdAmount
    );
    
    event AuctionEnded(
        uint256 indexed auctionId,
        address indexed winner,
        uint256 winningBid,
        address indexed seller
    );
    
    event AuctionCancelled(uint256 indexed auctionId);
    event PriceFeedUpdated(address token, address priceFeed);
    event FeeUpdated(uint256 feePercentage, address feeRecipient);
    
    // 初始化函数
    function initialize(
        address _ethUsdPriceFeed,
        address _feeRecipient
    ) public initializer {
        __Ownable_init(msg.sender);
        __UUPSUpgradeable_init();
        
        ethUsdPriceFeed = _ethUsdPriceFeed;
        feeRecipient = _feeRecipient;
        feePercentage = 250; // 2.5%
    }
    
    // UUPS升级授权
    function _authorizeUpgrade(address newImplementation) internal override onlyOwner {}
    
    // 设置价格Feed
    function setPriceFeed(address token, address priceFeed) external onlyOwner {
        priceFeeds[token] = priceFeed;
        emit PriceFeedUpdated(token, priceFeed);
    }
    
    // 设置手续费
    function setFee(uint256 _feePercentage, address _feeRecipient) external onlyOwner {
        require(_feePercentage <= 1000, "Fee too high"); // 最大10%
        feePercentage = _feePercentage;
        feeRecipient = _feeRecipient;
        emit FeeUpdated(_feePercentage, _feeRecipient);
    }
    
    // 创建拍卖
    function createAuction(
        address nftAddress,
        uint256 tokenId,
        uint256 duration, // 秒
        uint256 startPrice,
        address paymentToken
    ) external returns (uint256) {
        require(duration >= 1 hours, "Duration too short");
        require(duration <= 30 days, "Duration too long");
        require(startPrice > 0, "Start price must be > 0");
        
        // 确保NFT属于调用者
        IERC721 nft = IERC721(nftAddress);
        require(nft.ownerOf(tokenId) == msg.sender, "Not NFT owner");
        
        // 确保拍卖不存在
        require(nftToAuctionId[nftAddress][tokenId] == 0, "Auction already exists");
        
        // 转移NFT到合约
        nft.transferFrom(msg.sender, address(this), tokenId);
        
        // 创建拍卖
        uint256 auctionId = uint256(keccak256(abi.encodePacked(
            nftAddress,
            tokenId,
            block.timestamp,
            msg.sender
        )));
        
        Auction storage auction = auctions[auctionId];
        auction.id = auctionId;
        auction.seller = msg.sender;
        auction.nftAddress = nftAddress;
        auction.tokenId = tokenId;
        auction.startTime = block.timestamp;
        auction.endTime = block.timestamp + duration;
        auction.startPrice = startPrice;
        auction.paymentToken = paymentToken;
        auction.status = AuctionStatus.ACTIVE;
        
        nftToAuctionId[nftAddress][tokenId] = auctionId;
        
        emit AuctionCreated(
            auctionId,
            msg.sender,
            nftAddress,
            tokenId,
            auction.startTime,
            auction.endTime,
            startPrice,
            paymentToken
        );
        
        return auctionId;
    }
    
    // 出价
    function placeBid(uint256 auctionId, uint256 amount) external payable {
        Auction storage auction = auctions[auctionId];
        
        require(auction.status == AuctionStatus.ACTIVE, "Auction not active");
        require(block.timestamp < auction.endTime, "Auction ended");
        
        uint256 bidAmount;
        if (auction.paymentToken == address(0)) {
            // ETH出价
            bidAmount = msg.value;
            require(bidAmount >= auction.highestBid + auction.startPrice * 1 / 100, "Bid too low");
            require(bidAmount >= auction.startPrice, "Bid below start price");
        } else {
            // ERC20出价
            bidAmount = amount;
            require(bidAmount >= auction.highestBid + auction.startPrice * 1 / 100, "Bid too low");
            require(bidAmount >= auction.startPrice, "Bid below start price");
            
            IERC20 token = IERC20(auction.paymentToken);
            token.safeTransferFrom(msg.sender, address(this), bidAmount);
        }
        
        // 退还前一个最高出价
        if (auction.highestBidder != address(0)) {
            _refundBid(auctionId, auction.highestBidder, auction.highestBid);
        }
        
        // 更新最高出价
        auction.highestBidder = msg.sender;
        auction.highestBid = bidAmount;
        
        // 记录出价
        auctionBids[auctionId].push(Bid({
            bidder: msg.sender,
            amount: bidAmount,
            timestamp: block.timestamp,
            refunded: false
        }));
        
        // 计算美元价值
        uint256 usdAmount = getBidAmountInUSD(auction.paymentToken, bidAmount);
        
        emit BidPlaced(auctionId, msg.sender, bidAmount, usdAmount);
    }
    
    // 结束拍卖
    function endAuction(uint256 auctionId) external {
        Auction storage auction = auctions[auctionId];
        
        require(auction.status == AuctionStatus.ACTIVE, "Auction not active");
        require(block.timestamp >= auction.endTime || msg.sender == auction.seller, "Cannot end early");
        
        auction.status = AuctionStatus.ENDED;
        
        if (auction.highestBidder != address(0)) {
            // 转移NFT给获胜者
            IERC721 nft = IERC721(auction.nftAddress);
            nft.transferFrom(address(this), auction.highestBidder, auction.tokenId);
            
            // 计算手续费
            uint256 fee = auction.highestBid * feePercentage / 10000;
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
    
    // 取消拍卖
    function cancelAuction(uint256 auctionId) external {
        Auction storage auction = auctions[auctionId];
        
        require(msg.sender == auction.seller || msg.sender == owner(), "Not authorized");
        require(auction.status == AuctionStatus.ACTIVE, "Auction not active");
        require(auction.highestBidder == address(0), "Cannot cancel with bids");
        
        auction.status = AuctionStatus.CANCELLED;
        
        // 退还NFT
        IERC721 nft = IERC721(auction.nftAddress);
        nft.transferFrom(address(this), auction.seller, auction.tokenId);
        
        emit AuctionCancelled(auctionId);
        
        // 清理映射
        delete nftToAuctionId[auction.nftAddress][auction.tokenId];
    }
    
    // 获取美元价格
    function getBidAmountInUSD(address token, uint256 amount) public view returns (uint256) {
        address priceFeedAddress;
        
        if (token == address(0)) {
            priceFeedAddress = ethUsdPriceFeed;
        } else {
            priceFeedAddress = priceFeeds[token];
        }
        
        require(priceFeedAddress != address(0), "Price feed not set");
        
        AggregatorV3Interface priceFeed = AggregatorV3Interface(priceFeedAddress);
        (, int256 price, , , ) = priceFeed.latestRoundData();
        uint8 decimals = priceFeed.decimals();
        
        // 计算美元价值
        return (amount * uint256(price)) / (10 ** decimals);
    }
    
    // 获取拍卖详情
    function getAuction(uint256 auctionId) external view returns (Auction memory) {
        return auctions[auctionId];
    }
    
    // 获取拍卖出价记录
    function getAuctionBids(uint256 auctionId) external view returns (Bid[] memory) {
        return auctionBids[auctionId];
    }
    
    // 获取活跃拍卖
    function getActiveAuctions() external view returns (Auction[] memory) {
        uint256 count = 0;
        
        // 先计数
        for (uint256 i = 0; i < type(uint256).max; i++) {
            if (auctions[i].id != 0 && auctions[i].status == AuctionStatus.ACTIVE) {
                count++;
            }
        }
        
        // 填充数组
        Auction[] memory activeAuctions = new Auction[](count);
        uint256 index = 0;
        
        for (uint256 i = 0; i < type(uint256).max; i++) {
            if (auctions[i].id != 0 && auctions[i].status == AuctionStatus.ACTIVE) {
                activeAuctions[index] = auctions[i];
                index++;
            }
        }
        
        return activeAuctions;
    }
    
    // 内部函数：退还出价
    function _refundBid(uint256 auctionId, address bidder, uint256 amount) internal {
        Auction storage auction = auctions[auctionId];
        
        if (auction.paymentToken == address(0)) {
            payable(bidder).transfer(amount);
        } else {
            IERC20 token = IERC20(auction.paymentToken);
            token.safeTransfer(bidder, amount);
        }
        
        // 标记为已退还
        for (uint256 i = 0; i < auctionBids[auctionId].length; i++) {
            if (auctionBids[auctionId][i].bidder == bidder && 
                auctionBids[auctionId][i].amount == amount &&
                !auctionBids[auctionId][i].refunded) {
                auctionBids[auctionId][i].refunded = true;
                break;
            }
        }
    }
    
    // 接收ETH
    receive() external payable {}
}