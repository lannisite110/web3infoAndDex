import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useMemo, type ReactNode } from "react";
import { WagmiProvider } from "wagmi";
import { ErrorBoundary } from "../components/ErrorBoundary";
import { DexConfigProvider } from "./DexConfigContext";
import { createWagmiConfig } from "./createWagmiConfig";
import { resolveDexConfig } from "./resolveConfig";
import type { NFTAuctionDexConfig } from "./types";

const queryClient = new QueryClient();

type Props = {
  config?: NFTAuctionDexConfig;
  children: ReactNode;
};

/** Providers required by NFTAuctionDex (wagmi, react-query, contract config). */
export function NFTAuctionDexProvider({ config, children }: Props) {
  const resolved = useMemo(() => resolveDexConfig(config), [config]);
  const wagmiConfig = useMemo(
    () => createWagmiConfig(resolved.sepoliaRpcUrl),
    [resolved.sepoliaRpcUrl],
  );

  return (
    <WagmiProvider config={wagmiConfig}>
      <QueryClientProvider client={queryClient}>
        <DexConfigProvider config={config}>
          <ErrorBoundary>{children}</ErrorBoundary>
        </DexConfigProvider>
      </QueryClientProvider>
    </WagmiProvider>
  );
}
