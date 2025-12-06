import { describe, it, beforeEach } from "node:test";
import { expect } from "chai";
import hre from "hardhat";
import { parseEther, parseUnits } from "viem";
import { loadFixture } from "@nomicfoundation/hardhat-toolbox-viem/testing";

describe("Auction Upgrade", function () {
  async function deployContractsFixture() {
    const [owner, seller] = await hre.viem.getWalletClients();
    
    // 部署Mock价格Feed
    const MockPriceFeed = await hre.viem.deployContract("MockV3Aggregator", [8, 2000n * 10n ** 8n]);
    
    // 部署NFT合约
    const MyNFT = await hre.viem.deployContract("MyNFT", []);
    
    // 部署拍卖合约V1
    const AuctionV1 = await hre.viem.deployContract("AuctionV1", []);
    
    // 初始化拍卖合约
    await AuctionV1.write.initialize([MockPriceFeed.address, owner.account.address]);
    
    // 铸造NFT
    await MyNFT.write.safeMint([seller.account.address, "ipfs://test-nft-1"]);
    
    return { owner, seller, MyNFT, AuctionV1, MockPriceFeed };
  }
  
  describe("UUPS Upgrade", function () {
    it("Should upgrade to V2 successfully", async function () {
      const { owner, MyNFT, AuctionV1 } = await loadFixture(deployContractsFixture);
      
      // 授权NFT并创建拍卖
      await MyNFT.write.approve([AuctionV1.address, 0n], {
        account: owner.account,
      });
      
      const auctionId = await AuctionV1.write.createAuction([
        MyNFT.address,
        0n,
        3600n,
        parseEther("1"),
        "0x0000000000000000000000000000000000000000",
      ], {
        account: owner.account,
      });
      
      const auctionBefore = await AuctionV1.read.getAuction([auctionId]);
      
      // 部署AuctionV2
      const AuctionV2 = await hre.viem.getContractFactory("AuctionV2");
      const auctionV2Address = await hre.upgrades.upgradeProxy(
        AuctionV1.address,
        AuctionV2
      );
      
      // 重新获取合约实例
      const auctionV2 = await hre.viem.getContractAt(
        "AuctionV2",
        auctionV2Address
      );
      
      // 初始化V2功能
      await auctionV2.write.initializeV2();
      
      // 验证状态保持不变
      const auctionAfter = await auctionV2.read.getAuction([auctionId]);
      expect(auctionAfter[1]).to.equal(auctionBefore[1]); // seller
      expect(auctionAfter[2]).to.equal(auctionBefore[2]); // nftAddress
      expect(auctionAfter[3]).to.equal(auctionBefore[3]); // tokenId
      
      // 验证V2功能
      const version = await auctionV2.read.version();
      expect(version).to.equal("V2.0.0");
      
      const feeTiers = await auctionV2.read.getAllFeeTiers();
      expect(feeTiers.length).to.equal(4);
    });
    
    it("Should use dynamic fees in V2", async function () {
      const { owner, MyNFT } = await loadFixture(deployContractsFixture);
      
      // 部署新拍卖合约V2
      const AuctionV2 = await hre.viem.deployContract("AuctionV2", []);
      
      const MockPriceFeed = await hre.viem.deployContract("MockV3Aggregator", [8, 2000n * 10n ** 8n]);
      
      // 初始化
      await AuctionV2.write.initialize([MockPriceFeed.address, owner.account.address]);
      await AuctionV2.write.initializeV2();
      
      // 授权并创建拍卖
      await MyNFT.write.approve([AuctionV2.address, 0n], {
        account: owner.account,
      });
      
      const auctionId = await AuctionV2.write.createAuction([
        MyNFT.address,
        0n,
        3600n,
        parseEther("1"),
        "0x0000000000000000000000000000000000000000",
      ], {
        account: owner.account,
      });
      
      // 测试动态手续费
      const lowBid = parseEther("0.5"); // $1000
      const mediumBid = parseEther("5"); // $10000
      const highBid = parseEther("25"); // $50000
      
      const lowFee = await AuctionV2.read.getDynamicFee([1000n * 10n ** 18n]);
      const mediumFee = await AuctionV2.read.getDynamicFee([10000n * 10n ** 18n]);
      const highFee = await AuctionV2.read.getDynamicFee([50000n * 10n ** 18n]);
      
      expect(lowFee).to.equal(250n); // 2.5%
      expect(mediumFee).to.equal(200n); // 2.0%
      expect(highFee).to.equal(150n); // 1.5%
    });
  });
});