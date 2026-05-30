import { hasCustomRpc, rpcHint } from "../config/rpc";

export function RpcWarning() {
  if (hasCustomRpc) return null;

  return (
    <div className="alert" role="alert">
      <strong>RPC 未配置：</strong> 当前使用公共节点{" "}
      <code>rpc.sepolia.org</code>，容易失败（Failed to fetch）。
      <br />
      {rpcHint}
    </div>
  );
}
