import type { NFTAuctionDexConfig, ResolvedDexConfig } from "./types";

const ZERO = "0x0000000000000000000000000000000000000000" as const;
const DEFAULT_TEST_NFT = "0x8D8AD875810933D40dba91378c680d39223114c9" as `0x${string}`;
const DEFAULT_NFT_AUCTION =
  "0x751D5EDA4EFA561702EFfAe3d6096B28206df575" as `0x${string}`;
const DEFAULT_CHAIN_ID = 11155111;

function envAddress(name: "VITE_TEST_NFT_ADDRESS" | "VITE_NFT_AUCTION_ADDRESS") {
  const v = import.meta.env[name];
  return (typeof v === "string" && v.startsWith("0x") ? v : undefined) as
    | `0x${string}`
    | undefined;
}

export function resolveDexConfig(
  overrides?: NFTAuctionDexConfig,
): ResolvedDexConfig {
  const nftAuctionAddress =
    overrides?.nftAuctionAddress ??
    envAddress("VITE_NFT_AUCTION_ADDRESS") ??
    DEFAULT_NFT_AUCTION;

  const testNftAddress =
    overrides?.testNftAddress ??
    envAddress("VITE_TEST_NFT_ADDRESS") ??
    DEFAULT_TEST_NFT;

  const fromEnv = (import.meta.env.VITE_API_URL ?? "").replace(/\/$/, "");
  let apiBaseUrl = overrides?.apiBaseUrl ?? fromEnv;
  if (apiBaseUrl === undefined || apiBaseUrl === "") {
    apiBaseUrl = import.meta.env.PROD ? "" : "http://localhost:8080";
  }

  const sepoliaRpcUrl =
    overrides?.sepoliaRpcUrl ??
    import.meta.env.VITE_SEPOLIA_RPC_URL ??
    "https://rpc.sepolia.org";

  return {
    nftAuctionAddress,
    testNftAddress,
    apiBaseUrl,
    sepoliaRpcUrl,
    chainId: overrides?.chainId ?? DEFAULT_CHAIN_ID,
  };
}

export function isZeroAddress(addr: `0x${string}`): boolean {
  return addr.toLowerCase() === ZERO;
}

export function isApiEnabled(apiBaseUrl: string): boolean {
  return import.meta.env.PROD || apiBaseUrl.length > 0;
}
