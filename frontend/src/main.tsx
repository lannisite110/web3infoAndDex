import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import { NFTAuctionDexProvider } from "./dex";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <NFTAuctionDexProvider>
      <App />
    </NFTAuctionDexProvider>
  </StrictMode>,
);
