import { useState } from "react";
import { ConnectWallet } from "./components/ConnectWallet";
import { CreateAuctionForm } from "./components/CreateAuctionForm";
import { AuctionList } from "./components/AuctionList";
import {
  NFT_AUCTION_ADDRESS,
  TEST_NFT_ADDRESS,
} from "./config/contracts";

export default function App() {
  const [listKey, setListKey] = useState(0);
  const nftConfigured =
    TEST_NFT_ADDRESS !== "0x0000000000000000000000000000000000000000";

  return (
    <div className="app">
      <header>
        <h1>NFT 拍卖 · Sepolia</h1>
        <ConnectWallet />
      </header>

      {!nftConfigured && (
        <div className="alert">
          <strong>首次使用请先部署合约：</strong>
          <br />
          在项目根目录执行（需 Node 22+）：
          <br />
          <code>nvm use 22</code>
          <br />
          <code>npx hardhat run scripts/deploy-all.ts --network sepolia</code>
          <br />
          会自动更新 <code>frontend/src/config/contracts.ts</code>
        </div>
      )}

      <section className="card">
        <h2>合约地址</h2>
        <p className="muted">拍卖合约</p>
        <p style={{ wordBreak: "break-all", fontSize: "0.85rem" }}>
          {NFT_AUCTION_ADDRESS}
        </p>
        {nftConfigured && (
          <>
            <p className="muted">测试 NFT</p>
            <p style={{ wordBreak: "break-all", fontSize: "0.85rem" }}>
              {TEST_NFT_ADDRESS}
            </p>
          </>
        )}
      </section>

      <section className="card">
        <h2>创建拍卖</h2>
        <CreateAuctionForm onSuccess={() => setListKey((k) => k + 1)} />
      </section>

      <section>
        <h2>拍卖列表</h2>
        <AuctionList key={listKey} />
      </section>
    </div>
  );
}
