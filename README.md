# web3infoAndDex

Web3 学习项目：Sepolia 测试网上的 NFT 拍卖合约 + React / wagmi 前端。

## 结构

- `contracts/` — `NFTAuction`、`TestNFT`
- `scripts/` — Hardhat 部署脚本
- `frontend/` — Vite + React + wagmi 拍卖页面

## 部署合约（需 Node 22+）

```bash
nvm use 22
cp .env.example .env   # 填写 PRIVATE_KEY、ALCHEMY_API_URL
npm run deploy:all
```

## 本地前端

```bash
cd frontend
cp .env.example .env   # 填写 VITE_SEPOLIA_RPC_URL
npm install
npm run dev
```

## 测试

```bash
npx hardhat test
```

## 部署前端到 Vercel（免费）

1. 打开 [vercel.com](https://vercel.com)，用 **GitHub** 登录  
2. **Add New → Project**，导入 `lannisite110/web3infoAndDex`  
3. **Root Directory** 设为 `frontend`（必须）  
4. **Environment Variables** 添加：

   | 名称 | 值 |
   |------|-----|
   | `VITE_SEPOLIA_RPC_URL` | 与本地相同的 Sepolia RPC URL |
   | `VITE_NFT_AUCTION_ADDRESS` | `deploy-all` 输出的 NFTAuction 地址 |
   | `VITE_TEST_NFT_ADDRESS` | `deploy-all` 输出的 TestNFT 地址 |

5. 点 **Deploy**，完成后会得到 `https://xxx.vercel.app`  
6. MetaMask 选 **Sepolia** 后打开该链接即可使用

改环境变量后，在 Vercel 项目里 **Redeploy** 一次才会生效。
