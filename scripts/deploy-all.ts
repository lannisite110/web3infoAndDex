import { writeFileSync, mkdirSync, readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { network } from "hardhat";

const __dirname = dirname(fileURLToPath(import.meta.url));
const configPath = join(__dirname, "../frontend/src/config/contracts.ts");

async function deployAndMint(
  ethers: Awaited<ReturnType<typeof network.create>>["ethers"],
  deployer: { address: string },
) {
  console.log("1/2 部署 TestNFT...");
  const testNft = await ethers.deployContract("TestNFT");
  await testNft.waitForDeployment();
  const testNftAddress = await testNft.getAddress();
  console.log("   TestNFT:", testNftAddress);

  console.log("2/2 部署 NFTAuction...");
  const nftAuction = await ethers.deployContract("NFTAuction");
  await nftAuction.waitForDeployment();
  const nftAuctionAddress = await nftAuction.getAddress();
  console.log("   NFTAuction:", nftAuctionAddress);

  console.log("\n3/3 给自己 mint 3 个测试 NFT...");
  const testNftContract = await ethers.getContractAt("TestNFT", testNftAddress);
  for (let i = 0; i < 3; i++) {
    const tx = await testNftContract.mint(deployer.address);
    await tx.wait();
    console.log(`   mint #${i + 1} 完成`);
  }

  return { testNftAddress, nftAuctionAddress };
}

async function main() {
  console.log("=== 阶段 0：部署测试网合约 ===\n");

  const { ethers } = await network.create({ network: "sepolia" });
  const [deployer] = await ethers.getSigners();
  console.log("部署账户:", deployer.address);

  const balance = await ethers.provider.getBalance(deployer.address);
  console.log("账户余额:", ethers.formatEther(balance), "ETH\n");

  const { testNftAddress, nftAuctionAddress } = await deployAndMint(
    ethers,
    deployer,
  );

  mkdirSync(dirname(configPath), { recursive: true });
  const configContent = `// 由 scripts/deploy-all.ts 自动生成，请勿手改地址（可改 RPC）
export const SEPOLIA_CHAIN_ID = 11155111 as const;

export const TEST_NFT_ADDRESS =
  "${testNftAddress}" as const;

export const NFT_AUCTION_ADDRESS =
  "${nftAuctionAddress}" as const;
`;
  writeFileSync(configPath, configContent);
  console.log("\n已写入前端配置:", configPath);

  const abiDir = join(__dirname, "../frontend/src/abi");
  mkdirSync(abiDir, { recursive: true });
  for (const [name, artifact] of [
    ["NFTAuction", "contracts/NFTAuction.sol/NFTAuction.json"],
    ["TestNFT", "contracts/TestNFT.sol/TestNFT.json"],
  ] as const) {
    const json = JSON.parse(
      readFileSync(join(__dirname, "../artifacts", artifact), "utf8"),
    );
    writeFileSync(
      join(abiDir, `${name}.json`),
      JSON.stringify(json.abi, null, 2),
    );
  }

  console.log("\n✅ 阶段 0 完成。下一步：cd frontend && npm run dev");
  console.log("\nVercel 环境变量：");
  console.log("  VITE_TEST_NFT_ADDRESS=", testNftAddress);
  console.log("  VITE_NFT_AUCTION_ADDRESS=", nftAuctionAddress);
}

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
