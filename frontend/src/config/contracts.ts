// 本地默认值；Vercel 可在环境变量里覆盖
const DEFAULT_TEST_NFT = "0x61d3Ba91c1C12F376e8879d136F172A57BEEa5eA";
const DEFAULT_NFT_AUCTION = "0xF6E2DD42F7E6f37948B2a1A62AdE7B51f2018cEa";

export const SEPOLIA_CHAIN_ID = 11155111 as const;

export const TEST_NFT_ADDRESS = (import.meta.env.VITE_TEST_NFT_ADDRESS ??
  DEFAULT_TEST_NFT) as `0x${string}`;

export const NFT_AUCTION_ADDRESS = (import.meta.env.VITE_NFT_AUCTION_ADDRESS ??
  DEFAULT_NFT_AUCTION) as `0x${string}`;
