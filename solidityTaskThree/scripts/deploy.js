const { ethers } = require("hardhat");

async function main() {
  const [deployer] = await ethers.getSigners();
  console.log("Deployer:", deployer.address);

  // 获取合约工厂
  const MyNFT = await ethers.getContractFactory("MyNFT");
  
  // 部署合约
  const myNFT = await MyNFT.deploy();
  
  // 等待部署完成（注意：不是调用 deployed() 方法）
  await myNFT.waitForDeployment();
  
  console.log("MyNFT deployed to:", await myNFT.getAddress());
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });