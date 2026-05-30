import { useQuery } from "@tanstack/react-query";
import {
  fetchAuctionsFromApi,
  type AuctionQueryParams,
} from "../api/auctions";
import { isApiEnabled, isZeroAddress } from "../dex/resolveConfig";
import { useDexConfig } from "../dex/DexConfigContext";
import type { AuctionView } from "../types/auction";
import { useAuctionsFromChain } from "./useAuctionsFromChain";

export type AuctionDataSource = "api" | "chain";

/**
 * Auction list: prefers backend API when configured,
 * falls back to on-chain reads only after API retries are exhausted.
 */
export function useAuctions(searchParams: AuctionQueryParams = {}) {
  const { apiBaseUrl, nftAuctionAddress } = useDexConfig();
  const useApi = isApiEnabled(apiBaseUrl);

  const apiQuery = useQuery({
    queryKey: ["auctions", apiBaseUrl, searchParams],
    queryFn: () => fetchAuctionsFromApi(apiBaseUrl, searchParams),
    enabled: useApi,
    staleTime: 10_000,
    refetchInterval: 15_000,
    retry: 5,
    retryDelay: (attempt) => Math.min(3000 * 2 ** attempt, 30000),
  });

  const apiFailed = useApi && apiQuery.isFetched && apiQuery.isError;
  const apiPending = useApi && !apiQuery.isFetched && (apiQuery.isLoading || apiQuery.isFetching);
  const useChain = !useApi || apiFailed;

  const chain = useAuctionsFromChain(useChain);

  const source: AuctionDataSource =
    useApi && apiQuery.isSuccess ? "api" : "chain";

  let auctions =
    source === "api" && apiQuery.data ? apiQuery.data : chain.auctions;

  // Chain fallback: apply client-side filters when API unavailable
  if (source === "chain" && hasActiveFilters(searchParams)) {
    auctions = filterAuctionsClient(auctions, searchParams);
  }

  const isLoading =
    apiPending || (apiFailed && chain.isLoading) || (!useApi && chain.isLoading);

  const refetch = async () => {
    const tasks: Promise<unknown>[] = [];
    if (useApi) tasks.push(apiQuery.refetch());
    if (useChain) tasks.push(chain.refetch());
    await Promise.all(tasks);
  };

  return {
    auctions,
    isLoading,
    refetch,
    hasConfig: !isZeroAddress(nftAuctionAddress),
    source,
    apiPending,
    apiError: apiFailed ? apiQuery.error : null,
    apiBaseUrl,
  };
}

function hasActiveFilters(p: AuctionQueryParams): boolean {
  return Boolean(
    p.q?.trim() ||
      p.seller?.trim() ||
      p.tokenId?.trim() ||
      p.bidder?.trim() ||
      p.ended?.trim(),
  );
}

function filterAuctionsClient(auctions: AuctionView[], p: AuctionQueryParams) {
  return auctions.filter((a) => {
    if (p.seller?.trim() && !a.seller.toLowerCase().startsWith(p.seller.trim().toLowerCase())) {
      return false;
    }
    if (p.tokenId?.trim() && a.tokenId.toString() !== p.tokenId.trim()) {
      return false;
    }
    if (p.ended === "true" && !a.ended) return false;
    if (p.ended === "false" && a.ended) return false;
    if (p.q?.trim()) {
      const q = p.q.trim().toLowerCase();
      const match =
        a.seller.toLowerCase().includes(q) ||
        a.tokenId.toString() === q ||
        a.id.toString() === q;
      if (!match) return false;
    }
    return true;
  });
}
