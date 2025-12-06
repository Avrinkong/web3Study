import hre from "hardhat";

async function main() {
  const [deployer] = await hre.viem.getWalletClients();
  
  console.log("Verifying contracts on", hre.network.name);
  
  // 读取部署信息
  const deploymentPath = `deployments/${hre.network.name}/all-deployments.json`;
  const fs = await import("fs");
  
  if (!fs.existsSync(deploymentPath)) {
    console.log("No deployments found to verify");
    return;
  }
  
  const deployments = JSON.parse(fs.readFileSync(deploymentPath, "utf8"));
  
  for (const [contractName, info] of Object.entries(deployments)) {
    try {
      console.log(`Verifying ${contractName} at ${info.address}...`);
      
      await hre.run("verify:verify", {
        address: info.address,
        constructorArguments: info.args || [],
      });
      
      console.log(`✓ ${contractName} verified successfully`);
    } catch (error) {
      if (error.message.includes("Already Verified")) {
        console.log(`✓ ${contractName} already verified`);
      } else {
        console.log(`✗ Failed to verify ${contractName}:`, error.message);
      }
    }
  }
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});