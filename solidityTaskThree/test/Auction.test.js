const { expect } = require("chai");
const { ethers, upgrades } = require("hardhat");

describe("Auction Market", function () {
  let MyNFT;
  let myNFT;
  let AuctionV1;
  let auction;
  let owner;
  let seller;
  let bidder1;
  let bidder2;
  
  // Chainlink价格Feed模拟器
  let MockPriceFeed;
  let mockEthPriceFeed;
  let mockUsdcPriceFeed;
  
  // 测试用ERC20代币
  let MockERC20;
  let mockUSDC;

  beforeEach(async function () {
    [owner, seller, bidder1, bidder2] = await ethers.getSigners();
    
    // 部署Mock价格Feed
    MockPriceFeed = await ethers.getContractFactory("MockV3Aggregator");
    mockEthPriceFeed = await MockPriceFeed.deploy(8, 2000 * 10 ** 8); // $2000
    mockUsdcPriceFeed = await MockPriceFeed.deploy(8, 1 * 10 ** 8); // $1
    
    // 部署Mock ERC20代币
    MockERC20 = await ethers.getContractFactory("MockERC20");
    mockUSDC = await MockERC20.deploy("USDC", "USDC", 6);
    await mockUSDC.waitForDeployment();
    
    // 部署NFT合约
    MyNFT = await ethers.getContractFactory("MyNFT");
    myNFT = await MyNFT.deploy();
    await myNFT.waitForDeployment();
    
    // 部署拍卖合约V1
    AuctionV1 = await ethers.getContractFactory("AuctionV1");
    auction = await upgrades.deployProxy(
      AuctionV1,
      [await mockEthPriceFeed.getAddress(), owner.address],
      {
        initializer: "initialize",
        kind: "uups",
      }
    );
    await auction.waitForDeployment();
    
    // 设置USDC价格Feed
    await auction.setPriceFeed(await mockUSDC.getAddress(), await mockUsdcPriceFeed.getAddress());
    
    // 铸造NFT给卖家
    await myNFT.connect(owner).safeMint(seller.address, "ipfs://test-nft-1");
    
    // 给竞拍者一些ETH和USDC
    await mockUSDC.connect(owner).mint(bidder1.address, ethers.parseUnits("10000", 6));
    await mockUSDC.connect(owner).mint(bidder2.address, ethers.parseUnits("10000", 6));
  });

  describe("Deployment and Configuration", function () {
    it("Should set correct initial values", async function () {
      expect(await auction.ethUsdPriceFeed()).to.equal(await mockEthPriceFeed.getAddress());
      expect(await auction.feeRecipient()).to.equal(owner.address);
      expect(await auction.feePercentage()).to.equal(250); // 2.5%
    });

    it("Should allow owner to update price feeds", async function () {
      const newPriceFeed = await MockPriceFeed.deploy(8, 3000 * 10 ** 8);
      
      await auction.setPriceFeed(ethers.ZeroAddress, await newPriceFeed.getAddress());
      expect(await auction.priceFeeds(ethers.ZeroAddress)).to.equal(await newPriceFeed.getAddress());
    });

    it("Should allow owner to update fees", async function () {
      await auction.setFee(300, bidder1.address);
      
      expect(await auction.feePercentage()).to.equal(300);
      expect(await auction.feeRecipient()).to.equal(bidder1.address);
    });

    it("Should not allow non-owners to update settings", async function () {
      await expect(
        auction.connect(seller).setFee(100, seller.address)
      ).to.be.revertedWithCustomError(auction, "OwnableUnauthorizedAccount");
    });
  });

  describe("Auction Creation", function () {
    beforeEach(async function () {
      // 卖家授权NFT给拍卖合约
      await myNFT.connect(seller).approve(await auction.getAddress(), 0);
    });

    it("Should create auction successfully", async function () {
      const duration = 3600; // 1小时
      const startPrice = ethers.parseEther("1");
      
      await expect(
        auction.connect(seller).createAuction(
          await myNFT.getAddress(),
          0,
          duration,
          startPrice,
          ethers.ZeroAddress // ETH拍卖
        )
      ).to.emit(auction, "AuctionCreated");
      
      const auctionId = await auction.nftToAuctionId(await myNFT.getAddress(), 0);
      const auctionData = await auction.getAuction(auctionId);
      
      expect(auctionData.seller).to.equal(seller.address);
      expect(auctionData.nftAddress).to.equal(await myNFT.getAddress());
      expect(auctionData.tokenId).to.equal(0);
      expect(auctionData.startPrice).to.equal(startPrice);
      expect(auctionData.paymentToken).to.equal(ethers.ZeroAddress);
      expect(auctionData.status).to.equal(1); // ACTIVE
      
      // NFT应转移到拍卖合约
      expect(await myNFT.ownerOf(0)).to.equal(await auction.getAddress());
    });

    it("Should create USDC auction successfully", async function () {
      const duration = 3600;
      const startPrice = ethers.parseUnits("100", 6); // 100 USDC
      
      // 卖家授权USDC给拍卖合约
      await mockUSDC.connect(seller).approve(await auction.getAddress(), startPrice);
      
      await expect(
        auction.connect(seller).createAuction(
          await myNFT.getAddress(),
          0,
          duration,
          startPrice,
          await mockUSDC.getAddress()
        )
      ).to.emit(auction, "AuctionCreated");
      
      const auctionId = await auction.nftToAuctionId(await myNFT.getAddress(), 0);
      const auctionData = await auction.getAuction(auctionId);
      
      expect(auctionData.paymentToken).to.equal(await mockUSDC.getAddress());
    });

    it("Should not allow duplicate auctions", async function () {
      const duration = 3600;
      const startPrice = ethers.parseEther("1");
      
      await auction.connect(seller).createAuction(
        await myNFT.getAddress(),
        0,
        duration,
        startPrice,
        ethers.ZeroAddress
      );
      
      // 尝试创建重复拍卖
      await myNFT.connect(owner).safeMint(seller.address, "ipfs://test-nft-2");
      await myNFT.connect(seller).approve(await auction.getAddress(), 1);
      
      await expect(
        auction.connect(seller).createAuction(
          await myNFT.getAddress(),
          0, // 相同的tokenId
          duration,
          startPrice,
          ethers.ZeroAddress
        )
      ).to.be.revertedWith("Auction already exists");
    });

    it("Should validate auction duration", async function () {
      const shortDuration = 1800; // 30分钟，太短
      const longDuration = 31 * 24 * 3600; // 31天，太长
      const validDuration = 2 * 3600; // 2小时
      const startPrice = ethers.parseEther("1");
      
      await expect(
        auction.connect(seller).createAuction(
          await myNFT.getAddress(),
          0,
          shortDuration,
          startPrice,
          ethers.ZeroAddress
        )
      ).to.be.revertedWith("Duration too short");
      
      await expect(
        auction.connect(seller).createAuction(
          await myNFT.getAddress(),
          0,
          longDuration,
          startPrice,
          ethers.ZeroAddress
        )
      ).to.be.revertedWith("Duration too long");
    });
  });

  describe("Bidding", function () {
    let auctionId;
    const duration = 3600;
    const startPrice = ethers.parseEther("1");
    
    beforeEach(async function () {
      // 创建拍卖
      await myNFT.connect(seller).approve(await auction.getAddress(), 0);
      await auction.connect(seller).createAuction(
        await myNFT.getAddress(),
        0,
        duration,
        startPrice,
        ethers.ZeroAddress
      );
      
      auctionId = await auction.nftToAuctionId(await myNFT.getAddress(), 0);
      
      // 授权USDC
      await mockUSDC.connect(bidder1).approve(await auction.getAddress(), ethers.parseUnits("10000", 6));
      await mockUSDC.connect(bidder2).approve(await auction.getAddress(), ethers.parseUnits("10000", 6));
    });

    it("Should allow bidding with ETH", async function () {
      const bidAmount = ethers.parseEther("1.5");
      
      await expect(
        auction.connect(bidder1).placeBid(auctionId, 0, { value: bidAmount })
      ).to.emit(auction, "BidPlaced");
      
      const auctionData = await auction.getAuction(auctionId);
      expect(auctionData.highestBidder).to.equal(bidder1.address);
      expect(auctionData.highestBid).to.equal(bidAmount);
    });

    it("Should allow bidding with USDC", async function () {
      // 创建USDC拍卖
      await myNFT.connect(owner).safeMint(seller.address, "ipfs://test-nft-2");
      await myNFT.connect(seller).approve(await auction.getAddress(), 1);
      
      await auction.connect(seller).createAuction(
        await myNFT.getAddress(),
        1,
        duration,
        ethers.parseUnits("100", 6),
        await mockUSDC.getAddress()
      );
      
      const usdcAuctionId = await auction.nftToAuctionId(await myNFT.getAddress(), 1);
      const bidAmount = ethers.parseUnits("150", 6);
      
      await expect(
        auction.connect(bidder1).placeBid(usdcAuctionId, bidAmount)
      ).to.emit(auction, "BidPlaced");
      
      const auctionData = await auction.getAuction(usdcAuctionId);
      expect(auctionData.highestBidder).to.equal(bidder1.address);
      expect(auctionData.highestBid).to.equal(bidAmount);
    });

    it("Should require higher bid", async function () {
      const firstBid = ethers.parseEther("1.5");
      const sameBid = ethers.parseEther("1.5");
      const lowerBid = ethers.parseEther("1.4");
      const higherBid = ethers.parseEther("1.6");
      
      // 第一次出价
      await auction.connect(bidder1).placeBid(auctionId, 0, { value: firstBid });
      
      // 相同金额出价
      await expect(
        auction.connect(bidder2).placeBid(auctionId, 0, { value: sameBid })
      ).to.be.revertedWith("Bid too low");
      
      // 更低金额出价
      await expect(
        auction.connect(bidder2).placeBid(auctionId, 0, { value: lowerBid })
      ).to.be.revertedWith("Bid too low");
      
      // 更高金额出价
      await expect(
        auction.connect(bidder2).placeBid(auctionId, 0, { value: higherBid })
      ).to.emit(auction, "BidPlaced");
      
      // 检查前一个出价是否被退还
      const auctionData = await auction.getAuction(auctionId);
      expect(auctionData.highestBidder).to.equal(bidder2.address);
      expect(auctionData.highestBid).to.equal(higherBid);
    });

    it("Should not allow bidding after auction ends", async function () {
      // 快进时间到拍卖结束后
      await ethers.provider.send("evm_increaseTime", [duration + 1]);
      await ethers.provider.send("evm_mine");
      
      await expect(
        auction.connect(bidder1).placeBid(auctionId, 0, { value: startPrice })
      ).to.be.revertedWith("Auction ended");
    });
  });

  describe("Auction Ending", function () {
    let auctionId;
    const duration = 3600;
    const startPrice = ethers.parseEther("1");
    
    beforeEach(async function () {
      // 创建拍卖并出价
      await myNFT.connect(seller).approve(await auction.getAddress(), 0);
      await auction.connect(seller).createAuction(
        await myNFT.getAddress(),
        0,
        duration,
        startPrice,
        ethers.ZeroAddress
      );
      
      auctionId = await auction.nftToAuctionId(await myNFT.getAddress(), 0);
      
      // 多个出价
      await auction.connect(bidder1).placeBid(auctionId, 0, { value: ethers.parseEther("1.5") });
      await auction.connect(bidder2).placeBid(auctionId, 0, { value: ethers.parseEther("2.0") });
      await auction.connect(bidder1).placeBid(auctionId, 0, { value: ethers.parseEther("2.5") });
    });

    it("Should end auction successfully", async function () {
      // 快进时间到拍卖结束
      await ethers.provider.send("evm_increaseTime", [duration + 1]);
      await ethers.provider.send("evm_mine");
      
      const initialSellerBalance = await ethers.provider.getBalance(seller.address);
      const initialFeeRecipientBalance = await ethers.provider.getBalance(owner.address);
      
      await expect(auction.connect(owner).endAuction(auctionId))
        .to.emit(auction, "AuctionEnded")
        .withArgs(auctionId, bidder1.address, ethers.parseEther("2.5"), seller.address);
      
      // 检查NFT转移
      expect(await myNFT.ownerOf(0)).to.equal(bidder1.address);
      
      // 检查资金分配
      const finalSellerBalance = await ethers.provider.getBalance(seller.address);
      const finalFeeRecipientBalance = await ethers.provider.getBalance(owner.address);
      
      const fee = ethers.parseEther("2.5") * 250n / 10000n; // 2.5%手续费
      const sellerAmount = ethers.parseEther("2.5") - fee;
      
      // 注意：gas费用会影响余额，所以我们检查大致范围
      expect(finalSellerBalance - initialSellerBalance).to.be.closeTo(sellerAmount, ethers.parseEther("0.01"));
      expect(finalFeeRecipientBalance - initialFeeRecipientBalance).to.be.closeTo(fee, ethers.parseEther("0.01"));
    });

    it("Should allow seller to end auction early with no bids", async function () {
      // 创建新拍卖但不出价
      await myNFT.connect(owner).safeMint(seller.address, "ipfs://test-nft-2");
      await myNFT.connect(seller).approve(await auction.getAddress(), 1);
      
      await auction.connect(seller).createAuction(
        await myNFT.getAddress(),
        1,
        duration,
        startPrice,
        ethers.ZeroAddress
      );
      
      const newAuctionId = await auction.nftToAuctionId(await myNFT.getAddress(), 1);
      
      // 卖家提前结束
      await auction.connect(seller).endAuction(newAuctionId);
      
      // NFT应退还卖家
      expect(await myNFT.ownerOf(1)).to.equal(seller.address);
    });

    it("Should not allow ending active auction early by others", async function () {
      await expect(
        auction.connect(bidder1).endAuction(auctionId)
      ).to.be.revertedWith("Cannot end early");
    });

    it("Should cancel auction successfully", async function () {
      // 创建新拍卖但不出价
      await myNFT.connect(owner).safeMint(seller.address, "ipfs://test-nft-2");
      await myNFT.connect(seller).approve(await auction.getAddress(), 1);
      
      await auction.connect(seller).createAuction(
        await myNFT.getAddress(),
        1,
        duration,
        startPrice,
        ethers.ZeroAddress
      );
      
      const newAuctionId = await auction.nftToAuctionId(await myNFT.getAddress(), 1);
      
      await expect(auction.connect(seller).cancelAuction(newAuctionId))
        .to.emit(auction, "AuctionCancelled");
      
      const auctionData = await auction.getAuction(newAuctionId);
      expect(auctionData.status).to.equal(3); // CANCELLED
      expect(await myNFT.ownerOf(1)).to.equal(seller.address);
    });

    it("Should not cancel auction with bids", async function () {
      await expect(
        auction.connect(seller).cancelAuction(auctionId)
      ).to.be.revertedWith("Cannot cancel with bids");
    });
  });

  describe("Price Oracle", function () {
    it("Should get USD price for ETH bid", async function () {
      const ethAmount = ethers.parseEther("1");
      const usdAmount = await auction.getBidAmountInUSD(ethers.ZeroAddress, ethAmount);
      
      // $2000 * 1 ETH = $2000，考虑小数位
      const expectedAmount = 2000n * 10n ** 18n;
      expect(usdAmount).to.equal(expectedAmount);
    });

    it("Should get USD price for USDC bid", async function () {
      const usdcAmount = ethers.parseUnits("100", 6);
      const usdAmount = await auction.getBidAmountInUSD(await mockUSDC.getAddress(), usdcAmount);
      
      // $1 * 100 USDC = $100
      const expectedAmount = 100n * 10n ** 18n;
      expect(usdAmount).to.equal(expectedAmount);
    });
  });
});