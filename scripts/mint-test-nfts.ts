/**
 * 仅对已有 TestNFT 合约 mint，避免重复部署浪费 gas。
 * 用法：TEST_NFT_ADDRESS=0x... npx hardhat run scripts/mint-test-nfts.ts --network sepolia
 */
import { network } from "hardhat";

const DEFAULT_TEST_NFT = "0x61d3Ba91c1C12F376e8879d136F172A57BEEa5eA";

async function main() {
  const testNftAddress = process.env.TEST_NFT_ADDRESS ?? DEFAULT_TEST_NFT;

  const { ethers } = await network.create({ network: "sepolia" });
  const [deployer] = await ethers.getSigners();

  console.log("账户:", deployer.address);
  console.log("TestNFT:", testNftAddress);

  const testNft = await ethers.getContractAt("TestNFT", testNftAddress);

  for (let i = 0; i < 3; i++) {
    const tx = await testNft.mint(deployer.address);
    await tx.wait();
    console.log(`mint #${i + 1} 完成`);
  }

  console.log("\n✅ 已 mint 3 个 NFT，Token ID 一般为 1、2、3");
}

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
