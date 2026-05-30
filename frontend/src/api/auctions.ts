import type { AuctionView } from "../types/auction";

export type ApiAuction = {
  chainId: number;
  contract: string;
  auctionId: number;
  seller: string;
  nftContract: string;
  tokenId: string;
  startPrice: string;
  highestBid: string;
  highestBidder: string;
  startTime: number;
  duration: number;
  ended: boolean;
};

type ApiAuctionListResponse = {
  auctions: ApiAuction[];
};

export function mapApiAuction(a: ApiAuction): AuctionView {
  return {
    id: a.auctionId,
    seller: a.seller as `0x${string}`,
    nftContract: a.nftContract as `0x${string}`,
    tokenId: BigInt(a.tokenId),
    startPrice: BigInt(a.startPrice),
    startTime: BigInt(a.startTime),
    duration: BigInt(a.duration),
    highestBidder: a.highestBidder as `0x${string}`,
    highestBid: BigInt(a.highestBid),
    ended: a.ended,
  };
}

export async function fetchAuctionsFromApi(baseUrl: string): Promise<AuctionView[]> {
  const res = await fetch(`${baseUrl}/api/v1/auctions`);
  if (!res.ok) {
    throw new Error(`API ${res.status}: ${res.statusText}`);
  }
  const body = (await res.json()) as ApiAuctionListResponse;
  if (!Array.isArray(body.auctions)) {
    throw new Error("Invalid auctions response");
  }
  return body.auctions.map(mapApiAuction).sort((a, b) => a.id - b.id);
}
