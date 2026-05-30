import { useQuery } from "@tanstack/react-query";
import { fetchAuctionsFromApi } from "../api/auctions";
import { isApiEnabled, isZeroAddress } from "../dex/resolveConfig";
import { useDexConfig } from "../dex/DexConfigContext";
import { useAuctionsFromChain } from "./useAuctionsFromChain";

export type AuctionDataSource = "api" | "chain";

/**
 * Auction list: prefers backend API when configured,
 * falls back to on-chain reads only after API retries are exhausted.
 */
export function useAuctions() {
  const { apiBaseUrl, nftAuctionAddress } = useDexConfig();
  const useApi = isApiEnabled(apiBaseUrl);

  const apiQuery = useQuery({
    queryKey: ["auctions", apiBaseUrl],
    queryFn: () => fetchAuctionsFromApi(apiBaseUrl),
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

  const auctions =
    source === "api" && apiQuery.data ? apiQuery.data : chain.auctions;

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
