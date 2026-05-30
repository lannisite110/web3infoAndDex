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
