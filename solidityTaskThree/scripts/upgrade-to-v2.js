const { ethers, upgrades } = require("hardhat");

async function main() {
  const PROXY_ADDRESS = "YOUR_DEPLOYED_PROXY_ADDRESS"; // 替换为实际的代理地址

  const AuctionV2 = await ethers.getContractFactory("AuctionV2");
  console.log("Upgrading Auction to V2...");
  const upgraded = await upgrades.upgradeProxy(PROXY_ADDRESS, AuctionV2);
  console.log("Auction upgraded to V2 at:", upgraded.address);

  // 初始化V2的新功能
  console.log("Initializing V2 features...");
  await upgraded.initializeV2();
  console.log("V2 features initialized.");
}

main();