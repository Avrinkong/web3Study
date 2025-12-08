// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@chainlink/contracts/src/v0.8/interfaces/AggregatorV3Interface.sol";

contract AuctionV1 is Initializable, UUPSUpgradeable, OwnableUpgradeable {
    // 价格预言机映射，address(0)代表ETH
    mapping(address => AggregatorV3Interface) public priceFeeds;

    struct Auction {
        address seller;
        address nftContract;
        uint256 tokenId;
        uint256 startTime;
        uint256 endTime;
        uint256 startPrice;
        address paymentToken; // 支付代币地址，address(0) 代表 ETH
        address highestBidder;
        uint256 highestBid;
        bool ended;
    }
    mapping(uint256 => Auction) public auctions;
    uint256 public auctionCount;
    uint256 public feePercentage; // 手续费，100 = 1%
    address public feeRecipient;

    event AuctionCreated(uint256 indexed auctionId, address indexed seller, uint256 tokenId);
    event BidPlaced(uint256 indexed auctionId, address indexed bidder, uint256 amount, uint256 usdValue);
    event AuctionEnded(uint256 indexed auctionId, address indexed winner, uint256 amount);

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(address _ethUsdPriceFeed, address _feeRecipient) public initializer {
        __Ownable_init(msg.sender);
        __UUPSUpgradeable_init();
        // 设置ETH/USD预言机
        priceFeeds[address(0)] = AggregatorV3Interface(_ethUsdPriceFeed);
        feeRecipient = _feeRecipient;
        feePercentage = 250; // 默认2.5%
    }

    // UUPS升级授权，仅所有者可升级
    function _authorizeUpgrade(address newImplementation) internal override onlyOwner {}

    // 设置代币价格预言机
    function setTokenPriceFeed(address token, address priceFeed) external onlyOwner {
        priceFeeds[token] = AggregatorV3Interface(priceFeed);
    }

    // 创建拍卖
    function createAuction(
        address nftContract,
        uint256 tokenId,
        uint256 duration,
        uint256 startPrice,
        address paymentToken
    ) external {
        IERC721(nftContract).transferFrom(msg.sender, address(this), tokenId);
        auctionCount++;
        auctions[auctionCount] = Auction({
            seller: msg.sender,
            nftContract: nftContract,
            tokenId: tokenId,
            startTime: block.timestamp,
            endTime: block.timestamp + duration,
            startPrice: startPrice,
            paymentToken: paymentToken,
            highestBidder: address(0),
            highestBid: 0,
            ended: false
        });
        emit AuctionCreated(auctionCount, msg.sender, tokenId);
    }

    // 出价函数（支持ETH和ERC20）
    function placeBid(uint256 auctionId, uint256 amount) external payable {
        Auction storage auction = auctions[auctionId];
        require(block.timestamp < auction.endTime, "Auction ended");
        require(!auction.ended, "Auction already ended");

        uint256 bidAmount;
        if (auction.paymentToken == address(0)) {
            // ETH出价
            bidAmount = msg.value;
            require(bidAmount > auction.highestBid, "Bid too low");
            require(bidAmount >= auction.startPrice, "Bid below start price");
        } else {
            // ERC20出价
            bidAmount = amount;
            require(bidAmount > auction.highestBid, "Bid too low");
            require(bidAmount >= auction.startPrice, "Bid below start price");
            IERC20 token = IERC20(auction.paymentToken);
            require(token.transferFrom(msg.sender, address(this), bidAmount), "Transfer failed");
        }

        // 退还前一位最高出价者的资金
        if (auction.highestBidder != address(0)) {
            _refundBid(auction, auction.highestBidder, auction.highestBid);
        }

        auction.highestBidder = msg.sender;
        auction.highestBid = bidAmount;

        // 使用Chainlink计算并记录美元价值
        uint256 usdValue = getBidValueInUSD(auction.paymentToken, bidAmount);
        emit BidPlaced(auctionId, msg.sender, bidAmount, usdValue);
    }

    // 结束拍卖并结算
    function endAuction(uint256 auctionId) external virtual  {
        Auction storage auction = auctions[auctionId];
        require(block.timestamp >= auction.endTime, "Auction not ended");
        require(!auction.ended, "Auction already ended");
        require(msg.sender == auction.seller || msg.sender == owner(), "Not authorized");

        auction.ended = true;

        if (auction.highestBidder != address(0)) {
            // 计算手续费和卖家所得
            uint256 fee = (auction.highestBid * feePercentage) / 10000;
            uint256 sellerProceeds = auction.highestBid - fee;

            // 转移资金
            if (auction.paymentToken == address(0)) {
                // ETH
                payable(feeRecipient).transfer(fee);
                payable(auction.seller).transfer(sellerProceeds);
            } else {
                // ERC20
                IERC20 token = IERC20(auction.paymentToken);
                token.transfer(feeRecipient, fee);
                token.transfer(auction.seller, sellerProceeds);
            }
            // 转移NFT
            IERC721(auction.nftContract).transferFrom(address(this), auction.highestBidder, auction.tokenId);
            emit AuctionEnded(auctionId, auction.highestBidder, auction.highestBid);
        } else {
            // 无人出价，退回NFT
            IERC721(auction.nftContract).transferFrom(address(this), auction.seller, auction.tokenId);
        }
    }

    // ========== 核心：Chainlink价格计算函数 ==========
    function getBidValueInUSD(address token, uint256 amount) public view returns (uint256) {
        AggregatorV3Interface priceFeed = priceFeeds[token];
        require(address(priceFeed) != address(0), "Price feed not set");

        (, int256 price, , , ) = priceFeed.latestRoundData();
        uint8 decimals = priceFeed.decimals();
        // 计算逻辑：金额 * 价格 / (10^价格精度)
        // 注意：此公式假设 `amount` 的小数位与代币本身一致（如ETH为18位），需根据实际代币调整
        return (amount * uint256(price)) / (10 ** uint256(decimals));
    }

    // 内部退款函数
    function _refundBid(Auction storage auction, address to, uint256 amount) internal {
        if (auction.paymentToken == address(0)) {
            payable(to).transfer(amount);
        } else {
            IERC20 token = IERC20(auction.paymentToken);
            token.transfer(to, amount);
        }
    }
}