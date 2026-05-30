import { http, createConfig } from "wagmi";
import { sepolia } from "wagmi/chains";
import { injected } from "wagmi/connectors";

export function createWagmiConfig(sepoliaRpcUrl: string) {
  return createConfig({
    chains: [sepolia],
    connectors: [injected()],
    transports: {
      [sepolia.id]: http(sepoliaRpcUrl),
    },
  });
}
