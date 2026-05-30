import { useAccount, useConnect, useDisconnect } from "wagmi";

export function ConnectWallet() {
  const { address, isConnected, chain } = useAccount();
  const { connect, connectors, isPending } = useConnect();
  const { disconnect } = useDisconnect();

  const injected = connectors[0];

  if (isConnected && address) {
    return (
      <div>
        <button type="button" className="secondary" onClick={() => disconnect()}>
          断开
        </button>
        <p className="muted" style={{ marginTop: "0.5rem", marginBottom: 0 }}>
          {address.slice(0, 6)}…{address.slice(-4)}
          {chain ? ` · ${chain.name}` : ""}
        </p>
      </div>
    );
  }

  return (
    <button
      type="button"
      disabled={!injected || isPending}
      onClick={() => connect({ connector: injected })}
    >
      {isPending ? "连接中…" : "连接 MetaMask"}
    </button>
  );
}
