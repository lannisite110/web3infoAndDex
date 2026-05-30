import { useEffect, useState } from "react";
import { formatEther, parseEther } from "viem";
import {
  useAccount,
  useWriteContract,
  useWaitForTransactionReceipt,
} from "wagmi";
import nftAuctionAbi from "../abi/NFTAuction.json";
import { useDexConfig } from "../dex/DexConfigContext";
import type { AuctionView } from "../types/auction";

type Props = {
  auction: AuctionView;
  onUpdated?: () => void;
};

function formatTimeLeft(auction: AuctionView): string {
  const end = Number(auction.startTime + auction.duration);
  const now = Math.floor(Date.now() / 1000);
  const left = end - now;
  if (left <= 0) return "已结束";
  const m = Math.floor(left / 60);
  const s = left % 60;
  return `${m} 分 ${s} 秒`;
}

export function AuctionCard({ auction, onUpdated }: Props) {
  const { nftAuctionAddress } = useDexConfig();
  const { address, isConnected } = useAccount();
  const [bidEth, setBidEth] = useState("0.02");

  const { writeContract, data: txHash, isPending, error, reset } =
    useWriteContract();

  const { isLoading: confirming, isSuccess } = useWaitForTransactionReceipt({
    hash: txHash,
  });

  useEffect(() => {
    if (isSuccess && txHash) {
      reset();
      onUpdated?.();
    }
  }, [isSuccess, txHash, reset, onUpdated]);

  const now = BigInt(Math.floor(Date.now() / 1000));
  const endedOnChain = auction.ended;
  const timeUp = now >= auction.startTime + auction.duration;
  const canEnd = !endedOnChain && timeUp;
  const canBid =
    !endedOnChain &&
    !timeUp &&
    isConnected &&
    address?.toLowerCase() !== auction.seller.toLowerCase();

  const minNextBid =
    auction.highestBid > 0n
      ? auction.highestBid + parseEther("0.001")
      : auction.startPrice;

  function bid() {
    const value = parseEther(bidEth);
    if (value < minNextBid) return;
    writeContract({
      address: nftAuctionAddress,
      abi: nftAuctionAbi,
      functionName: "bid",
      args: [BigInt(auction.id)],
      value,
    });
  }

  function endAuction() {
    writeContract({
      address: nftAuctionAddress,
      abi: nftAuctionAbi,
      functionName: "endAuction",
      args: [BigInt(auction.id)],
    });
  }

  const busy = isPending || confirming;

  return (
    <article className="card">
      <p>
        <strong>拍卖 #{auction.id}</strong>
        {endedOnChain ? " · 已结算" : timeUp ? " · 待结算" : " · 进行中"}
      </p>
      <p className="muted">卖家 {auction.seller.slice(0, 10)}…</p>
      <p>Token ID: {auction.tokenId.toString()}</p>
      <p>起拍: {formatEther(auction.startPrice)} ETH</p>
      <p>
        当前最高: {formatEther(auction.highestBid)} ETH
        {auction.highestBidder !== "0x0000000000000000000000000000000000000000"
          ? ` (${auction.highestBidder.slice(0, 8)}…)`
          : " (暂无出价)"}
      </p>
      <p>剩余: {formatTimeLeft(auction)}</p>

      {canBid && (
        <div style={{ marginTop: "0.75rem" }}>
          <label htmlFor={`bid-${auction.id}`}>出价 (ETH)</label>
          <input
            id={`bid-${auction.id}`}
            value={bidEth}
            onChange={(e) => setBidEth(e.target.value)}
          />
          <p className="muted">至少 {formatEther(minNextBid)} ETH</p>
          <button type="button" disabled={busy} onClick={bid}>
            {busy ? "提交中…" : "出价"}
          </button>
        </div>
      )}

      {canEnd && (
        <div className="row">
          <button type="button" disabled={busy} onClick={endAuction}>
            {busy ? "提交中…" : "结束并结算"}
          </button>
        </div>
      )}

      {error && (
        <p className="muted" style={{ color: "#f28b82" }}>
          {error.shortMessage ?? error.message}
        </p>
      )}
    </article>
  );
}
