/** Optional overrides; unset fields fall back to Vite env / built-in Sepolia defaults. */
export type NFTAuctionDexConfig = {
  nftAuctionAddress?: `0x${string}`;
  testNftAddress?: `0x${string}`;
  /** API base URL. Production embed: leave empty to use same-origin `/api` (Vercel rewrite). */
  apiBaseUrl?: string;
  sepoliaRpcUrl?: string;
  chainId?: number;
};

export type ResolvedDexConfig = {
  nftAuctionAddress: `0x${string}`;
  testNftAddress: `0x${string}`;
  apiBaseUrl: string;
  sepoliaRpcUrl: string;
  chainId: number;
};

export type NFTAuctionDexProps = {
  /** When true, wraps Wagmi + React Query + config (for iframe / single-script embed). */
  embedded?: boolean;
  config?: NFTAuctionDexConfig;
  title?: string;
  showHeader?: boolean;
  showContractInfo?: boolean;
  showCreateAuction?: boolean;
  /** Show REST API search form above auction list (phase 4a). */
  showAuctionSearch?: boolean;
  className?: string;
  onAuctionCreated?: () => void;
};
