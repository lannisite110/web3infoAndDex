import { useState } from "react";
import type { AuctionQueryParams } from "../api/auctions";
import { useAuctions } from "../hooks/useAuctions";
import { AuctionCard } from "./AuctionCard";
import { AuctionSearch } from "./AuctionSearch";

export function AuctionList() {
  const [searchParams, setSearchParams] = useState<AuctionQueryParams>({});
  const { auctions, isLoading, refetch, hasConfig, source, apiError, apiPending, apiBaseUrl } =
    useAuctions(searchParams);

  if (!hasConfig) {
    return <p className="muted">请配置 NFT_AUCTION_ADDRESS。</p>;
  }

  return (
    <div>
      <AuctionSearch onSearch={setSearchParams} />

      {isLoading ? (
        <p className="muted">
          {apiPending
            ? "正在连接后端 API（Render 冷启动可能需要 30～60 秒）…"
            : source === "api"
              ? "从 API 加载拍卖列表…"
              : "从链上加载拍卖列表…"}
        </p>
      ) : auctions.length === 0 ? (
        <p className="muted">没有匹配的拍卖。可调整搜索条件或创建新拍卖。</p>
      ) : (
        <>
          <div className="row" style={{ marginBottom: "0.75rem" }}>
            <button type="button" className="secondary" onClick={() => refetch()}>
              刷新列表
            </button>
            <span className="muted" style={{ fontSize: "0.85rem" }}>
              数据来源：{source === "api" ? "后端 API" : "链上 RPC"}
              {apiError ? (
                <>
                  （API 不可用，已回退链上
                  {apiError instanceof Error ? `：${apiError.message}` : ""}
                  ；API：{apiBaseUrl || "同源 /api"}）
                </>
              ) : null}
            </span>
          </div>
          {auctions.map((a) => (
            <AuctionCard key={a.id} auction={a} onUpdated={() => refetch()} />
          ))}
        </>
      )}
    </div>
  );
}
