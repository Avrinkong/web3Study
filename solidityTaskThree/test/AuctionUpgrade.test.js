const { expect } = require("chai");
const { ethers, upgrades } = require("hardhat");

describe("Auction Market Upgrade", function () {
  let MyNFT;
  let myNFT;
  let AuctionV1;
  let AuctionV2;
  let auction;
  let owner;
  let seller;
  let bidder1;
  
  beforeEach(async function () {
    [owner, seller, bidder1] = await ethers.getSigners();
    
    // 部署Mock价格Feed
    const MockPriceFeed = await ethers.getContractFactory("MockV3Aggregator");
    const mockEthPriceFeed = await MockPriceFeed.deploy(8, 2000 * 10 ** 8);
    
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
    
    // 铸造NFT
    await myNFT.connect(owner).safeMint(seller.address, "ipfs://test-nft-1");
  });

  describe("UUPS Upgrade", function () {
    it("Should upgrade to V2 successfully", async function () {
      // 部署V2合约
      AuctionV2 = await ethers.getContractFactory("AuctionV2");
      const auctionV2 = await upgrades.upgradeProxy(await auction.getAddress(), AuctionV2);
      await auctionV2.waitForDeployment();
      
      // 初始化V2功能
      await auctionV2.initializeV2();
      
      // 验证V2功能
      expect(await auctionV2.version()).to.equal("V2.0.0");
      
      // 检查手续费层级
      const feeTiers = await auctionV2.getAllFeeTiers();
      expect(feeTiers.length).to.equal(4);
      expect(feeTiers[0].minAmount).to.equal(0);
      expect(feeTiers[0].feePercentage).to.equal(250);
    });

    it("Should maintain state after upgrade", async function () {
      // 在升级前创建拍卖
      await myNFT.connect(seller).approve(await auction.getAddress(), 0);
      await auction.connect(seller).createAuction(
        await myNFT.getAddress(),
        0,
        3600,
        ethers.parseEther("1"),
        ethers.ZeroAddress
      );
      
      const auctionId = await auction.nftToAuctionId(await myNFT.getAddress(), 0);
      const auctionDataBefore = await auction.getAuction(auctionId);
      
      // 升级到V2
      AuctionV2 = await ethers.getContractFactory("AuctionV2");
      const auctionV2 = await upgrades.upgradeProxy(await auction.getAddress(), AuctionV2);
      await auctionV2.waitForDeployment();
      await auctionV2.initializeV2();
      
      // 验证状态保持不变
      const auctionDataAfter = await auctionV2.getAuction(auctionId);
      
      expect(auctionDataAfter.seller).to.equal(auctionDataBefore.seller);
      expect(auctionDataAfter.nftAddress).to.equal(auctionDataBefore.nftAddress);
      expect(auctionDataAfter.tokenId).to.equal(auctionDataBefore.tokenId);
      expect(auctionDataAfter.startPrice).to.equal(auctionDataBefore.startPrice);
      expect(auctionDataAfter.status).to.equal(auctionDataBefore.status);
    });

    it("Should use dynamic fees in V2", async function () {
      // 升级到V2
      AuctionV2 = await ethers.getContractFactory("AuctionV2");
      const auctionV2 = await upgrades.upgradeProxy(await auction.getAddress(), AuctionV2);
      await auctionV2.waitForDeployment();
      await auctionV2.initializeV2();
      
      // 创建拍卖
      await myNFT.connect(seller).approve(await auctionV2.getAddress(), 0);
      await auctionV2.connect(seller).createAuction(
        await myNFT.getAddress(),
        0,
        3600,
        ethers.parseEther("1"),
        ethers.ZeroAddress
      );
      
      const auctionId = await auctionV2.nftToAuctionId(await myNFT.getAddress(), 0);
      
      // 模拟高价出价（$5000）
      const highBid = ethers.parseEther("2.5"); // 2.5 ETH * $2000 = $5000
      await auctionV2.connect(bidder1).placeBid(auctionId, 0, { value: highBid });
      
      // 快进时间并结束拍卖
      await ethers.provider.send("evm_increaseTime", [3601]);
      await ethers.provider.send("evm_mine");
      
      const initialSellerBalance = await ethers.provider.getBalance(seller.address);
      await auctionV2.connect(owner).endAuction(auctionId);
      
      // 检查手续费：$5000应使用2.0%的手续费层级
      const finalSellerBalance = await ethers.provider.getBalance(seller.address);
      const received = finalSellerBalance - initialSellerBalance;
      
      const expectedFee = highBid * 200n / 10000n; // 2.0%
      const expectedSellerAmount = highBid - expectedFee;
      
      // 考虑到gas费用，检查大致范围
      expect(received).to.be.closeTo(expectedSellerAmount, ethers.parseEther("0.01"));
    });

    it("Should allow managing fee tiers", async function () {
      // 升级到V2
      AuctionV2 = await ethers.getContractFactory("AuctionV2");
      const auctionV2 = await upgrades.upgradeProxy(await auction.getAddress(), AuctionV2);
      await auctionV2.waitForDeployment();
      await auctionV2.initializeV2();
      
      // 添加新手续费层级
      await auctionV2.addFeeTier(50000n * 10n ** 18n, 50); // $50000以上0.5%
      
      const feeTiers = await auctionV2.getAllFeeTiers();
      expect(feeTiers.length).to.equal(5);
      
      // 移除手续费层级
      await auctionV2.removeFeeTier(2);
      
      const updatedFeeTiers = await auctionV2.getAllFeeTiers();
      expect(updatedFeeTiers.length).to.equal(4);
    });
  });
});