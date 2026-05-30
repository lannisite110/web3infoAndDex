import type { AuctionView } from "../types/auction";
import type { BidRecord } from "../types/bid";

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

export type AuctionQueryParams = {
  q?: string;
  seller?: string;
  tokenId?: string;
  ended?: string;
  bidder?: string;
};

type ApiAuctionListResponse = {
  auctions: ApiAuction[];
};

type ApiBidsResponse = {
  auctionId: number;
  bids: BidRecord[];
};

function fetchWithTimeout(url: string, ms: number): Promise<Response> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), ms);
  return fetch(url, { method: "GET", signal: controller.signal }).finally(() =>
    clearTimeout(timer),
  );
}

function buildUrl(baseUrl: string, path: string, params?: AuctionQueryParams): string {
  const fullPath = baseUrl ? `${baseUrl}${path}` : path;
  if (!params) return fullPath;
  const sp = new URLSearchParams();
  if (params.q?.trim()) sp.set("q", params.q.trim());
  if (params.seller?.trim()) sp.set("seller", params.seller.trim());
  if (params.tokenId?.trim()) sp.set("tokenId", params.tokenId.trim());
  if (params.ended?.trim()) sp.set("ended", params.ended.trim());
  if (params.bidder?.trim()) sp.set("bidder", params.bidder.trim());
  const qs = sp.toString();
  return qs ? `${fullPath}?${qs}` : fullPath;
}

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

export async function fetchAuctionsFromApi(
  baseUrl: string,
  params?: AuctionQueryParams,
): Promise<AuctionView[]> {
  const url = buildUrl(baseUrl, "/api/v1/auctions", params);
  const res = await fetchWithTimeout(url, 90_000);
  if (!res.ok) {
    throw new Error(`API ${res.status}: ${res.statusText}`);
  }
  const body = (await res.json()) as ApiAuctionListResponse;
  if (!Array.isArray(body.auctions)) {
    throw new Error("Invalid auctions response");
  }
  return body.auctions.map(mapApiAuction).sort((a, b) => a.id - b.id);
}

export async function fetchBidsForAuction(
  baseUrl: string,
  auctionId: number,
): Promise<BidRecord[]> {
  const path = `/api/v1/auctions/${auctionId}/bids`;
  const url = baseUrl ? `${baseUrl}${path}` : path;
  const res = await fetchWithTimeout(url, 30_000);
  if (!res.ok) {
    throw new Error(`API ${res.status}: ${res.statusText}`);
  }
  const body = (await res.json()) as ApiBidsResponse;
  if (!Array.isArray(body.bids)) {
    throw new Error("Invalid bids response");
  }
  return body.bids;
}
