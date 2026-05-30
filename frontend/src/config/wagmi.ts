import { http, createConfig } from "wagmi";
import { sepolia } from "wagmi/chains";
import { injected } from "wagmi/connectors";

const rpcUrl =
  import.meta.env.VITE_SEPOLIA_RPC_URL ?? "https://rpc.sepolia.org";

/** wagmi：统一管理链、钱包连接器、RPC */
export const wagmiConfig = createConfig({
  chains: [sepolia],
  connectors: [injected()],
  transports: {
    [sepolia.id]: http(rpcUrl),
  },
});
