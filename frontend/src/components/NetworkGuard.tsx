import { useAccount, useSwitchChain } from "wagmi";
import { sepolia } from "wagmi/chains";
import { SEPOLIA_CHAIN_ID } from "../config/contracts";

export function NetworkGuard() {
  const { isConnected, chainId } = useAccount();
  const { switchChain, isPending } = useSwitchChain();

  if (!isConnected || chainId === SEPOLIA_CHAIN_ID) {
    return null;
  }

  return (
    <div className="alert" role="alert">
      <strong>网络不对：</strong>
      当前是链 ID {chainId}，请切换到 <strong>Sepolia</strong>。
      MetaMask 在错误网络上常只显示「查看提醒」且无法确认。
      <div className="row" style={{ marginTop: "0.75rem" }}>
        <button
          type="button"
          disabled={isPending}
          onClick={() => switchChain({ chainId: sepolia.id })}
        >
          {isPending ? "切换中…" : "切换到 Sepolia"}
        </button>
      </div>
    </div>
  );
}
