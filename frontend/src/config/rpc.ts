/** 是否配置了可靠 RPC（Vercel 必须设 VITE_SEPOLIA_RPC_URL） */
export const hasCustomRpc = Boolean(
  import.meta.env.VITE_SEPOLIA_RPC_URL?.trim(),
);

export const rpcHint =
  "请在 Vercel → Settings → Environment Variables 添加 VITE_SEPOLIA_RPC_URL（与本地 Infura 地址相同），然后 Redeploy。";
