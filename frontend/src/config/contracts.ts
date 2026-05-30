// 本地默认值；Vercel 可在环境变量里覆盖（无需改代码重新部署地址）
const DEFAULT_TEST_NFT = "0x0000000000000000000000000000000000000000";
const DEFAULT_NFT_AUCTION = "0x5b763484dabB5f857D246f922Ca1c34361EbB9e5";

export const SEPOLIA_CHAIN_ID = 11155111 as const;

export const TEST_NFT_ADDRESS = (import.meta.env.VITE_TEST_NFT_ADDRESS ??
  DEFAULT_TEST_NFT) as `0x${string}`;

export const NFT_AUCTION_ADDRESS = (import.meta.env.VITE_NFT_AUCTION_ADDRESS ??
  DEFAULT_NFT_AUCTION) as `0x${string}`;
