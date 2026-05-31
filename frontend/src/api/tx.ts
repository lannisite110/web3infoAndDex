export type TxReceipt = {
  txHash: string;
  status: string;
  blockNumber: string;
  from: string;
  to: string;
  gasUsed: string;
  etherscanUrl: string;
};

export function etherscanTxUrl(txHash: string): string {
  return `https://sepolia.etherscan.io/tx/${txHash}`;
}

export async function fetchTxReceipt(
  baseUrl: string,
  txHash: string,
): Promise<TxReceipt> {
  const path = `/api/v1/tx/${txHash}`;
  const url = baseUrl ? `${baseUrl}${path}` : path;
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`API ${res.status}`);
  }
  const body = (await res.json()) as { transaction: TxReceipt };
  return body.transaction;
}
