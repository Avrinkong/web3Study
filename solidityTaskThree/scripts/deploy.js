// scripts/deploy.js
const { ethers } = require("hardhat");

async function main() {
  const [deployer] = await ethers.getSigners();
  console.log("Deployer:", deployer.address);
  // 已移除有问题的 getBalance 调用

  // 1. 获取合约工厂并部署
  const MyNFT = await ethers.getContractFactory("MyNFT");
  const myNFT = await MyNFT.deploy();
  
  // 等待部署完成
  await myNFT.waitForDeployment();
  
  const contractAddress = await myNFT.getAddress();
  console.log("MyNFT deployed to:", contractAddress);
  
  // 2. 等待区块确认
  console.log("Waiting for block confirmations...");
  const deploymentReceipt = await myNFT.deploymentTransaction().wait(5);
  console.log(`Confirmed in block: ${deploymentReceipt.blockNumber}`);
  
  // 3. 自动验证合约
  console.log("Starting contract verification...");
  
  try {
    // 运行 Hardhat 的验证任务
    // 注意：你的 MyNFT 构造函数为空，所以 constructorArguments 为空数组
    await hre.run("verify:verify", {
      address: contractAddress,
      constructorArguments: [], // 关键：无参构造函数必须用空数组
    });
    console.log("✅ Contract successfully verified on Etherscan!");
  } catch (error) {
    // 处理常见的验证错误
    if (error.message.toLowerCase().includes("already verified")) {
      console.log("ℹ️ Contract is already verified.");
    } else {
      console.error("❌ Verification failed:", error.message);
      // 提供手动验证指令
      console.log("\n💡 You can try to verify manually later:");
      console.log(`npx hardhat verify --network sepolia ${contractAddress}`);
    }
  }
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error("Script failed:", error);
    process.exit(1);
  });