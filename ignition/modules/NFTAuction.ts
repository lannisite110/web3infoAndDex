import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("NFTAuctionModule", (m) => {
  const nftAuction = m.contract("NFTAuction");

  return { nftAuction };
});