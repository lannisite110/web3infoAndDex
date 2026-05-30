import { network } from "hardhat";

async function main() {
  console.log("开始部署 NFT 拍卖合约...");

  const { ethers } = await network.create({ network: "sepolia" });

  const [deployer] = await ethers.getSigners();

  console.log("部署账户：", deployer.address);

  const nftAuction = await ethers.deployContract("NFTAuction");

  console.log("✅ 部署完成！地址：", await nftAuction.getAddress());
}

main().catch((err) => {
  console.error("部署失败：", err);
  process.exitCode = 1;
});