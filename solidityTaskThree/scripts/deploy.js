const { ethers, upgrades } = require("hardhat");

async function main() {
  console.log("Starting deployment...");
  
  // Chainlink价格Feed地址（Sepolia测试网）
  const ETH_USD_PRICE_FEED = "0x694AA1769357215DE4FAC081bf1f309aDC325306"; // Sepolia ETH/USD
  const USDC_USD_PRICE_FEED = "0xA2F78ab2355fe2f984D808B5CeE7FD0A93D5270E"; // Sepolia USDC/USD
  
  const [deployer] = await ethers.getSigners();
  console.log("Deploying contracts with account:", deployer.address);
  
  // 1. 部署NFT合约
  console.log("Deploying MyNFT...");
  const MyNFT = await ethers.getContractFactory("MyNFT");
  const myNFT = await MyNFT.deploy();
  await myNFT.waitForDeployment();
  console.log("MyNFT deployed to:", await myNFT.getAddress());
  
  // 2. 部署拍卖合约V1
  console.log("Deploying AuctionV1...");
  const AuctionV1 = await ethers.getContractFactory("AuctionV1");
  const auctionV1 = await upgrades.deployProxy(
    AuctionV1,
    [ETH_USD_PRICE_FEED, deployer.address],
    {
      initializer: "initialize",
      kind: "uups",
    }
  );
  await auctionV1.waitForDeployment();
  console.log("AuctionV1 deployed to:", await auctionV1.getAddress());
  
  // 3. 设置USDC价格Feed
  console.log("Setting USDC price feed...");
  // 注意：这里需要USDC代币地址，Sepolia测试网USDC地址
  const USDC_ADDRESS = "0x94a9D9AC8a22534E3FaCa9F4e7F2E2cf85d5E4C8"; // Sepolia USDC
  await auctionV1.setPriceFeed(USDC_ADDRESS, USDC_USD_PRICE_FEED);
  
  console.log("Deployment completed!");
  console.log("=====================================");
  console.log("NFT Contract Address:", await myNFT.getAddress());
  console.log("Auction Contract Address:", await auctionV1.getAddress());
  console.log("Deployer Address:", deployer.address);
  console.log("=====================================");
  
  // 保存部署信息到文件
  const fs = require("fs");
  const deploymentInfo = {
    network: "sepolia",
    timestamp: new Date().toISOString(),
    contracts: {
      MyNFT: await myNFT.getAddress(),
      AuctionV1: await auctionV1.getAddress(),
    },
    deployer: deployer.address,
    priceFeeds: {
      ETH_USD: ETH_USD_PRICE_FEED,
      USDC_USD: USDC_USD_PRICE_FEED,
    },
  };
  
  fs.writeFileSync(
    "deployment-info.json",
    JSON.stringify(deploymentInfo, null, 2)
  );
  console.log("Deployment info saved to deployment-info.json");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });