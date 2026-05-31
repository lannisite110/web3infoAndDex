CREATE TABLE IF NOT EXISTS auctions (
  chain_id        BIGINT NOT NULL,
  contract        VARCHAR(42) NOT NULL,
  auction_id      BIGINT UNSIGNED NOT NULL,
  seller          VARCHAR(42) NOT NULL,
  nft_contract    VARCHAR(42) NOT NULL,
  token_id        VARCHAR(78) NOT NULL,
  start_price     VARCHAR(78) NOT NULL,
  highest_bid     VARCHAR(78) NOT NULL,
  highest_bidder  VARCHAR(42) NOT NULL,
  start_time      BIGINT NOT NULL,
  duration        BIGINT NOT NULL,
  ended           TINYINT(1) NOT NULL DEFAULT 0,
  updated_at      DATETIME(3) NOT NULL,
  PRIMARY KEY (chain_id, contract, auction_id),
  INDEX idx_seller (seller),
  INDEX idx_ended (ended),
  INDEX idx_token (nft_contract, token_id)
);

CREATE TABLE IF NOT EXISTS bids (
  chain_id      BIGINT NOT NULL,
  contract      VARCHAR(42) NOT NULL,
  auction_id    BIGINT UNSIGNED NOT NULL,
  bidder        VARCHAR(42) NOT NULL,
  amount        VARCHAR(78) NOT NULL,
  tx_hash       VARCHAR(66) NOT NULL,
  log_index     INT UNSIGNED NOT NULL,
  block_number  BIGINT UNSIGNED NOT NULL,
  indexed_at    DATETIME(3) NOT NULL,
  PRIMARY KEY (chain_id, contract, auction_id, tx_hash, log_index),
  INDEX idx_auction (chain_id, contract, auction_id),
  INDEX idx_bidder (bidder)
);

CREATE TABLE IF NOT EXISTS indexer_state (
  chain_id   BIGINT NOT NULL,
  contract   VARCHAR(42) NOT NULL,
  last_block BIGINT UNSIGNED NOT NULL,
  updated_at DATETIME(3) NOT NULL,
  PRIMARY KEY (chain_id, contract)
);
