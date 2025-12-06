import hre from "hardhat";
import { upgrades } from "@openzeppelin/hardhat-upgrades";

async function main() {
  const [deployer] = await hre.viem.getWalletClients();
  
  console.log("Upgrading Auction contract with account:", deployer.account.address);
  
  // 读取已部署的代理地址
  const AUCTION_PROXY_ADDRESS = "0x..."; // 替换为实际代理地址
  
  const AuctionV2 = await hre.viem.getContractFactory("AuctionV2");
  
  // 升级代理合约
  const auctionV2 = await upgrades.upgradeProxy(
    AUCTION_PROXY_ADDRESS,
    AuctionV2,
    { 
      deployer: deployer.account.address,
      kind: "uups"
    }
  );
  
  console.log("Auction upgraded to V2 at:", auctionV2.address);
  
  // 初始化V2功能
  const tx = await auctionV2.write.initializeV2();
  await hre.viem.waitForTransactionReceipt({ hash: tx });
  
  console.log("V2 features initialized");
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});