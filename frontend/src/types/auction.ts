/** 与合约 getAuction 返回结构对应 */
export type AuctionOnChain = {
  seller: `0x${string}`;
  nftContract: `0x${string}`;
  tokenId: bigint;
  startPrice: bigint;
  startTime: bigint;
  duration: bigint;
  highestBidder: `0x${string}`;
  highestBid: bigint;
  ended: boolean;
};

export type AuctionView = AuctionOnChain & {
  id: number;
};
