import { useAuctions } from "../hooks/useAuctions";
import { AuctionCard } from "./AuctionCard";

export function AuctionList() {
  const { auctions, isLoading, refetch, hasConfig } = useAuctions();

  if (!hasConfig) {
    return <p className="muted">请配置 NFT_AUCTION_ADDRESS。</p>;
  }

  if (isLoading) {
    return <p className="muted">从链上加载拍卖列表…</p>;
  }

  if (auctions.length === 0) {
    return (
      <p className="muted">
        还没有拍卖。连接钱包后可在上方创建第一场拍卖。
      </p>
    );
  }

  return (
    <div>
      <div className="row" style={{ marginBottom: "0.75rem" }}>
        <button type="button" className="secondary" onClick={() => refetch()}>
          刷新列表
        </button>
      </div>
      {auctions.map((a) => (
        <AuctionCard key={a.id} auction={a} onUpdated={() => refetch()} />
      ))}
    </div>
  );
}
