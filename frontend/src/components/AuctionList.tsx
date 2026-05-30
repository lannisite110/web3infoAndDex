import { useAuctions } from "../hooks/useAuctions";
import { AuctionCard } from "./AuctionCard";

export function AuctionList() {
  const { auctions, isLoading, refetch, hasConfig, source, apiError } =
    useAuctions();

  if (!hasConfig) {
    return <p className="muted">请配置 NFT_AUCTION_ADDRESS。</p>;
  }

  if (isLoading) {
    return (
      <p className="muted">
        {source === "api" ? "从 API 加载拍卖列表…" : "从链上加载拍卖列表…"}
      </p>
    );
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
        <span className="muted" style={{ fontSize: "0.85rem" }}>
          数据来源：{source === "api" ? "后端 API" : "链上 RPC"}
          {apiError ? "（API 不可用，已回退链上）" : ""}
        </span>
      </div>
      {auctions.map((a) => (
        <AuctionCard key={a.id} auction={a} onUpdated={() => refetch()} />
      ))}
    </div>
  );
}
