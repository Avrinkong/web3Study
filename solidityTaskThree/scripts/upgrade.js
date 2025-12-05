const { ethers, upgrades } = require("hardhat");

async function main() {
  console.log("Starting upgrade...");
  
  const [deployer] = await ethers.getSigners();
  console.log("Upgrading with account:", deployer.address);
  
  // 读取已部署的合约地址
  const fs = require("fs");
  const deploymentInfo = JSON.parse(fs.readFileSync("deployment-info.json", "utf8"));
  
  const auctionV1Address = deploymentInfo.contracts.AuctionV1;
  console.log("Current Auction address:", auctionV1Address);
  
  // 部署新版本的合约
  console.log("Deploying AuctionV2...");
  const AuctionV2 = await ethers.getContractFactory("AuctionV2");
  const auctionV2 = await upgrades.upgradeProxy(auctionV1Address, AuctionV2);
  await auctionV2.waitForDeployment();
  
  console.log("Auction upgraded to V2 at:", await auctionV2.getAddress());
  
  // 初始化新版本的功能
  console.log("Initializing V2 features...");
  await auctionV2.initializeV2();
  
  console.log("Upgrade completed!");
  
  // 更新部署信息
  deploymentInfo.contracts.AuctionV2 = await auctionV2.getAddress();
  deploymentInfo.upgradedAt = new Date().toISOString();
  
  fs.writeFileSync(
    "deployment-info-upgraded.json",
    JSON.stringify(deploymentInfo, null, 2)
  );
  console.log("Updated deployment info saved to deployment-info-upgraded.json");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });