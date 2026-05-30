import { useState } from "react";
import type { AuctionQueryParams } from "../api/auctions";

type Props = {
  onSearch: (params: AuctionQueryParams) => void;
};

const empty: AuctionQueryParams = {};

export function AuctionSearch({ onSearch }: Props) {
  const [q, setQ] = useState("");
  const [seller, setSeller] = useState("");
  const [tokenId, setTokenId] = useState("");
  const [bidder, setBidder] = useState("");
  const [ended, setEnded] = useState("");

  function submit(e: React.FormEvent) {
    e.preventDefault();
    onSearch({
      q: q.trim() || undefined,
      seller: seller.trim() || undefined,
      tokenId: tokenId.trim() || undefined,
      bidder: bidder.trim() || undefined,
      ended: ended || undefined,
    });
  }

  return (
    <form className="card" onSubmit={submit} style={{ marginBottom: "1rem" }}>
      <h3 style={{ marginTop: 0, fontSize: "1rem" }}>搜索拍卖（REST API）</h3>
      <div className="row" style={{ flexWrap: "wrap", gap: "0.5rem" }}>
        <label style={{ flex: "1 1 140px" }}>
          关键词
          <input
            placeholder="拍卖ID / Token / 卖家"
            value={q}
            onChange={(e) => setQ(e.target.value)}
          />
        </label>
        <label style={{ flex: "1 1 140px" }}>
          卖家地址
          <input
            placeholder="0x..."
            value={seller}
            onChange={(e) => setSeller(e.target.value)}
          />
        </label>
        <label style={{ flex: "0 1 100px" }}>
          Token ID
          <input value={tokenId} onChange={(e) => setTokenId(e.target.value)} />
        </label>
        <label style={{ flex: "1 1 140px" }}>
          出价人
          <input
            placeholder="0x..."
            value={bidder}
            onChange={(e) => setBidder(e.target.value)}
          />
        </label>
        <label style={{ flex: "0 1 120px" }}>
          状态
          <select value={ended} onChange={(e) => setEnded(e.target.value)}>
            <option value="">全部</option>
            <option value="false">进行中</option>
            <option value="true">已结束</option>
          </select>
        </label>
      </div>
      <div className="row" style={{ marginTop: "0.75rem" }}>
        <button type="submit">搜索</button>
        <button
          type="button"
          className="secondary"
          onClick={() => {
            setQ("");
            setSeller("");
            setTokenId("");
            setBidder("");
            setEnded("");
            onSearch(empty);
          }}
        >
          清除
        </button>
      </div>
    </form>
  );
}
