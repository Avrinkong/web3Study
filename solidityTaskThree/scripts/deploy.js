const { ethers, upgrades } = require("hardhat");
require("dotenv").config();

async function main() {
  const [deployer] = await ethers.getSigners();
  console.log("Deployer:", deployer.address);

  // 部署NFT合约
  const MyNFT = await ethers.getContractFactory("MyNFT");
  const myNFT = await MyNFT.deploy();
  await myNFT.deployed();
  console.log("MyNFT deployed to:", myNFT.address);

  // 部署可升级拍卖合约 (UUPS代理)
  const AuctionV1 = await ethers.getContractFactory("AuctionV1");
  const auctionProxy = await upgrades.deployProxy(
    AuctionV1,
    [
      "0x694AA1769357215DE4FAC081bf1f309aDC325306", // Sepolia ETH/USD 预言机
      deployer.address // 手续费接收地址
    ],
    { initializer: "initialize", kind: "uups" }
  );
  await auctionProxy.deployed();
  console.log("Auction Proxy deployed to:", auctionProxy.address);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});