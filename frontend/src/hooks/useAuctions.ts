import { useMemo } from "react";
import { useReadContract, useReadContracts } from "wagmi";
import nftAuctionAbi from "../abi/NFTAuction.json";
import { NFT_AUCTION_ADDRESS } from "../config/contracts";
import type { AuctionOnChain, AuctionView } from "../types/auction";

function parseAuction(raw: readonly unknown[]): AuctionOnChain {
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

/** 从链上读取拍卖总数 + 每条拍卖详情 */
export function useAuctions() {
  const { data: count, isLoading: countLoading, refetch: refetchCount } =
    useReadContract({
      address: NFT_AUCTION_ADDRESS,
      abi: nftAuctionAbi,
      functionName: "auctionCount",
    });

  const auctionIds = useMemo(() => {
    const n = count ? Number(count) : 0;
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
      query: { enabled: auctionIds.length > 0 },
    });

  const auctions: AuctionView[] = useMemo(() => {
    if (!results) return [];
    return results.flatMap((item, index) => {
      if (item.status !== "success" || !item.result) return [];
      const id = auctionIds[index];
      return [{ id, ...parseAuction(item.result as readonly unknown[]) }];
    });
  }, [results, auctionIds]);

  const refetch = async () => {
    await refetchCount();
    await refetchList();
  };

  return {
    auctions,
    isLoading: countLoading || listLoading,
    refetch,
    hasConfig: NFT_AUCTION_ADDRESS !== "0x0000000000000000000000000000000000000000",
  };
}
