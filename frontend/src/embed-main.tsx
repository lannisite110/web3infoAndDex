import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { NFTAuctionDex } from "./dex";

/**
 * Minimal embed demo: mount with embedded providers (iframe / third-party page).
 * Production: https://web3info-and-dex.vercel.app/embed.html
 */
createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <NFTAuctionDex
      embedded
      title="NFT 拍卖 DEX"
      showContractInfo={false}
      className="app"
    />
  </StrictMode>,
);
