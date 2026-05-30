import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { formatEther } from "viem";
import { fetchBidsForAuction } from "../api/auctions";
import { useDexConfig } from "../dex/DexConfigContext";
import { isApiEnabled } from "../dex/resolveConfig";

type Props = {
  auctionId: number;
};

export function BidHistory({ auctionId }: Props) {
  const { apiBaseUrl } = useDexConfig();
  const [open, setOpen] = useState(false);
  const canUseApi = isApiEnabled(apiBaseUrl);

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ["bids", apiBaseUrl, auctionId],
    queryFn: () => fetchBidsForAuction(apiBaseUrl, auctionId),
    enabled: open && canUseApi,
  });

  if (!canUseApi) {
    return (
      <p className="muted" style={{ fontSize: "0.85rem" }}>
        出价历史需后端 API（本地设 VITE_API_URL 或 Vercel 同源 /api）。
      </p>
    );
  }

  return (
    <div style={{ marginTop: "0.75rem" }}>
      <button
        type="button"
        className="secondary"
        onClick={() => setOpen((v) => !v)}
      >
        {open ? "收起出价历史" : "查看出价历史"}
      </button>
      {open && (
        <div style={{ marginTop: "0.5rem" }}>
          {isLoading && <p className="muted">加载中…</p>}
          {isError && (
            <p className="muted" style={{ color: "#f28b82" }}>
              加载失败，请确认后端已部署 Phase 4a。
              <button
                type="button"
                className="secondary"
                style={{ marginLeft: "0.5rem" }}
                onClick={() => refetch()}
              >
                重试
              </button>
            </p>
          )}
          {data && data.length === 0 && (
            <p className="muted">暂无出价记录（索引器会回填历史 Bid 事件）。</p>
          )}
          {data && data.length > 0 && (
            <ul className="bid-history">
              {data.map((b) => (
                <li key={`${b.txHash}-${b.logIndex}`}>
                  <span>{formatEther(BigInt(b.amount))} ETH</span>
                  <span className="muted">
                    {" "}
                    · {b.bidder.slice(0, 8)}… · 区块 {b.blockNumber}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  );
}
