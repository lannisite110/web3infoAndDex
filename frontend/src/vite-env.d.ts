/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_SEPOLIA_RPC_URL?: string;
  readonly VITE_TEST_NFT_ADDRESS?: string;
  readonly VITE_NFT_AUCTION_ADDRESS?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
