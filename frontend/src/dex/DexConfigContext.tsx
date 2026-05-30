import { createContext, useContext, useMemo, type ReactNode } from "react";
import { resolveDexConfig } from "./resolveConfig";
import type { NFTAuctionDexConfig, ResolvedDexConfig } from "./types";

const DexConfigContext = createContext<ResolvedDexConfig | null>(null);

type Props = {
  config?: NFTAuctionDexConfig;
  children: ReactNode;
};

export function DexConfigProvider({ config, children }: Props) {
  const value = useMemo(() => resolveDexConfig(config), [config]);
  return (
    <DexConfigContext.Provider value={value}>{children}</DexConfigContext.Provider>
  );
}

export function useDexConfig(): ResolvedDexConfig {
  const ctx = useContext(DexConfigContext);
  if (!ctx) {
    throw new Error("useDexConfig must be used within DexConfigProvider or NFTAuctionDex");
  }
  return ctx;
}
