/** @deprecated Prefer useDexConfig() — kept for deploy scripts writing this file. */
import { resolveDexConfig } from "../dex/resolveConfig";

const cfg = resolveDexConfig();

export const SEPOLIA_CHAIN_ID = cfg.chainId as 11155111;
export const TEST_NFT_ADDRESS = cfg.testNftAddress;
export const NFT_AUCTION_ADDRESS = cfg.nftAuctionAddress;
