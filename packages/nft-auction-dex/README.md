# @web3info/nft-auction-dex

可嵌入的 Sepolia NFT 拍卖 DEX React 组件（阶段 3）。

源码位于 monorepo 的 `frontend/src/dex/`，与演示站共用同一套 UI。

## 快速嵌入（iframe）

部署前端后，在任意页面嵌入：

```html
<iframe
  src="https://web3info-and-dex.vercel.app/embed.html"
  title="NFT Auction DEX"
  width="100%"
  height="900"
  style="border: 0; border-radius: 12px; max-width: 720px;"
></iframe>
```

`embed.html` 使用 `NFTAuctionDex` 的 `embedded` 模式（自带 Wagmi + API）。

## 在 React 项目中使用

### 1. 复制或引用模块

将 `frontend/src/dex/` 目录复制到你的项目，或在本仓库中通过相对路径导入。

### 2. 安装 peer 依赖

```bash
npm install react react-dom wagmi viem @tanstack/react-query
```

### 3. 挂载组件

**方式 A — 自带 Provider（推荐嵌入第三方站点）**

```tsx
import { NFTAuctionDex } from "./dex";

export function MyPage() {
  return (
    <NFTAuctionDex
      embedded
      title="我的拍卖"
      showContractInfo={false}
      config={{
        nftAuctionAddress: "0x751D5EDA4EFA561702EFfAe3d6096B28206df575",
        testNftAddress: "0x8D8AD875810933D40dba91378c680d39223114c9",
        apiBaseUrl: "", // 同源 /api 代理；或 https://your-api.onrender.com
        sepoliaRpcUrl: "https://sepolia.infura.io/v3/YOUR_KEY",
      }}
    />
  );
}
```

**方式 B — 外层已有 Wagmi**

```tsx
import { NFTAuctionDexProvider, NFTAuctionDex } from "./dex";

<NFTAuctionDexProvider config={{ ... }}>
  <NFTAuctionDex showHeader={false} />
</NFTAuctionDexProvider>
```

### 4. 样式

组件会加载 `index.css`（深色主题）。若需隔离样式，可包一层容器并覆盖 CSS 变量。

## 配置说明

| 字段 | 说明 |
|------|------|
| `nftAuctionAddress` | NFTAuction 合约 |
| `testNftAddress` | ERC721 合约 |
| `apiBaseUrl` | 后端 API；生产留空走 `/api` 代理 |
| `sepoliaRpcUrl` | Sepolia JSON-RPC |

环境变量（Vite）：`VITE_NFT_AUCTION_ADDRESS`、`VITE_TEST_NFT_ADDRESS`、`VITE_SEPOLIA_RPC_URL`、`VITE_API_URL`（本地开发）。

## 导出 API

```ts
export {
  NFTAuctionDex,
  NFTAuctionDexProvider,
  useDexConfig,
  resolveDexConfig,
  createWagmiConfig,
} from "./dex";
```
