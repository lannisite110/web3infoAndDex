export type BidRecord = {
  chainId: number;
  contract: string;
  auctionId: number;
  bidder: string;
  amount: string;
  txHash: string;
  logIndex: number;
  blockNumber: number;
  indexedAt: string;
};
