import { useMemo } from "react";
import { useReadContract, useReadContracts } from "wagmi";
import nftAuctionAbi from "../abi/NFTAuction.json";
import { NFT_AUCTION_ADDRESS } from "../config/contracts";
import type { AuctionOnChain, AuctionView } from "../types/auction";

type AuctionStruct = {
  seller: `0x${string}`;
  nftContract: `0x${string}`;
  tokenId: bigint;
  startPrice: bigint;
  startTime: bigint;
  duration: bigint;
  highestBidder: `0x${string}`;
  highestBid: bigint;
  ended: boolean;
};

function parseAuction(raw: unknown): AuctionOnChain | null {
  if (!raw || typeof raw !== "object") return null;

  if (Array.isArray(raw)) {
    const [
      seller,
      nftContract,
      tokenId,
      startPrice,
      startTime,
      duration,
      highestBidder,
      highestBid,
      ended,
    ] = raw;
    return {
      seller: seller as `0x${string}`,
      nftContract: nftContract as `0x${string}`,
      tokenId: tokenId as bigint,
      startPrice: startPrice as bigint,
      startTime: startTime as bigint,
      duration: duration as bigint,
      highestBidder: highestBidder as `0x${string}`,
      highestBid: highestBid as bigint,
      ended: ended as boolean,
    };
  }

  const s = raw as AuctionStruct;
  if (s.seller === undefined || s.tokenId === undefined) return null;
  return {
    seller: s.seller,
    nftContract: s.nftContract,
    tokenId: s.tokenId,
    startPrice: s.startPrice,
    startTime: s.startTime,
    duration: s.duration,
    highestBidder: s.highestBidder,
    highestBid: s.highestBid,
    ended: s.ended,
  };
}

/** Reads auction list directly from the NFTAuction contract. */
export function useAuctionsFromChain(enabled = true) {
  const { data: count, isLoading: countLoading, refetch: refetchCount } =
    useReadContract({
      address: NFT_AUCTION_ADDRESS,
      abi: nftAuctionAbi,
      functionName: "auctionCount",
      query: { enabled },
    });

  const auctionIds = useMemo(() => {
    const n = count ? Number(count) : 0;
    if (!Number.isFinite(n) || n < 0) return [];
    return Array.from({ length: n }, (_, i) => i + 1);
  }, [count]);

  const { data: results, isLoading: listLoading, refetch: refetchList } =
    useReadContracts({
      contracts: auctionIds.map((id) => ({
        address: NFT_AUCTION_ADDRESS,
        abi: nftAuctionAbi,
        functionName: "getAuction" as const,
        args: [BigInt(id)] as const,
      })),
      query: { enabled: enabled && auctionIds.length > 0 },
    });

  const auctions: AuctionView[] = useMemo(() => {
    if (!results) return [];
    return results.flatMap((item, index) => {
      if (item.status !== "success" || item.result == null) return [];
      const parsed = parseAuction(item.result);
      if (!parsed) return [];
      const id = auctionIds[index];
      return [{ id, ...parsed }];
    });
  }, [results, auctionIds]);

  const refetch = async () => {
    await refetchCount();
    await refetchList();
  };

  return {
    auctions,
    isLoading: enabled && (countLoading || listLoading),
    refetch,
    hasConfig:
      NFT_AUCTION_ADDRESS !== "0x0000000000000000000000000000000000000000",
  };
}
