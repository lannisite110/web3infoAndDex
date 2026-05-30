import { useState } from "react";
import { ConnectWallet } from "../components/ConnectWallet";
import { NetworkGuard } from "../components/NetworkGuard";
import { RpcWarning } from "../components/RpcWarning";
import { CreateAuctionForm } from "../components/CreateAuctionForm";
import { AuctionList } from "../components/AuctionList";
import { useDexConfig } from "./DexConfigContext";
import { NFTAuctionDexProvider } from "./NFTAuctionDexProvider";
import { isZeroAddress } from "./resolveConfig";
import type { NFTAuctionDexProps } from "./types";
import "../index.css";

function NFTAuctionDexInner({
  title = "NFT 拍卖 · Sepolia",
  showHeader = true,
  showContractInfo = true,
  showCreateAuction = true,
  showAuctionSearch = true,
  className = "app",
  onAuctionCreated,
}: Omit<NFTAuctionDexProps, "embedded" | "config">) {
  const { nftAuctionAddress, testNftAddress } = useDexConfig();
  const [listKey, setListKey] = useState(0);
  const nftConfigured = !isZeroAddress(testNftAddress);

  const bumpList = () => {
    setListKey((k) => k + 1);
    onAuctionCreated?.();
  };

  return (
    <div className={className}>
      {showHeader && (
        <header>
          <h1>{title}</h1>
          <ConnectWallet />
        </header>
      )}

      <NetworkGuard />
      <RpcWarning />

      {!nftConfigured && (
        <div className="alert">
          <strong>未配置 TestNFT 合约地址。</strong>
          请设置 <code>testNftAddress</code> 或环境变量{" "}
          <code>VITE_TEST_NFT_ADDRESS</code>。
        </div>
      )}

      {showContractInfo && (
        <section className="card">
          <h2>合约地址</h2>
          <p className="muted">拍卖合约</p>
          <p style={{ wordBreak: "break-all", fontSize: "0.85rem" }}>
            {nftAuctionAddress}
          </p>
          {nftConfigured && (
            <>
              <p className="muted">测试 NFT</p>
              <p style={{ wordBreak: "break-all", fontSize: "0.85rem" }}>
                {testNftAddress}
              </p>
            </>
          )}
        </section>
      )}

      {showCreateAuction && (
        <section className="card">
          <h2>创建拍卖</h2>
          <CreateAuctionForm onSuccess={bumpList} />
        </section>
      )}

      <section>
        <h2>拍卖列表</h2>
        <AuctionList key={listKey} />
      </section>
    </div>
  );
}

/**
 * Embeddable NFT auction DEX UI (Sepolia).
 * Set `embedded` when mounting without an outer WagmiProvider.
 */
export function NFTAuctionDex({
  embedded = false,
  config,
  ...ui
}: NFTAuctionDexProps) {
  const inner = <NFTAuctionDexInner {...ui} />;

  if (embedded) {
    return (
      <NFTAuctionDexProvider config={config}>{inner}</NFTAuctionDexProvider>
    );
  }

  return inner;
}
