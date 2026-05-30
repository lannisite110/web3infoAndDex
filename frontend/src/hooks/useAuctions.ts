import { useQuery } from "@tanstack/react-query";
import { fetchAuctionsFromApi } from "../api/auctions";
import { API_BASE_URL, isApiConfigured } from "../config/api";
import { useAuctionsFromChain } from "./useAuctionsFromChain";

export type AuctionDataSource = "api" | "chain";

/**
 * Auction list: prefers backend API when VITE_API_URL is set,
 * falls back to on-chain reads if the API is unavailable.
 */
export function useAuctions() {
  const useApi = isApiConfigured();

  const apiQuery = useQuery({
    queryKey: ["auctions", API_BASE_URL],
    queryFn: () => fetchAuctionsFromApi(API_BASE_URL),
    enabled: useApi,
    staleTime: 10_000,
    refetchInterval: 15_000,
    retry: 3,
    retryDelay: (attempt) => Math.min(1000 * 2 ** attempt, 15000),
  });

  const useChain = !useApi || apiQuery.isError;
  const chain = useAuctionsFromChain(useChain);

  const source: AuctionDataSource =
    useApi && !apiQuery.isError ? "api" : "chain";

  const auctions =
    source === "api" && apiQuery.data ? apiQuery.data : chain.auctions;

  const isLoading = useApi
    ? apiQuery.isLoading || (apiQuery.isError && chain.isLoading)
    : chain.isLoading;

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
    hasConfig: chain.hasConfig,
    source,
    apiError: useApi && apiQuery.isError ? apiQuery.error : null,
  };
}
