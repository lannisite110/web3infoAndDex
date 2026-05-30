/** 安全解析 Token ID，避免 BigInt("abc") 导致整页崩溃 */
export function parseTokenId(input: string): bigint | null {
  const trimmed = input.trim();
  if (!/^\d+$/.test(trimmed)) return null;
  try {
    return BigInt(trimmed);
  } catch {
    return null;
  }
}
