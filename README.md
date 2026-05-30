# web3infoAndDex

Web3 学习项目：Sepolia 测试网上的 NFT 拍卖合约 + React / wagmi 前端。

## 结构

- `contracts/` — `NFTAuction`、`TestNFT`
- `scripts/` — Hardhat 部署脚本
- `frontend/` — Vite + React + wagmi 拍卖页面
- `backend/` — Go API（阶段 2，Gin + MongoDB 索引）

## 本地后端 API（Go，阶段 2.1+）

### 1. MongoDB Atlas（免费 M0）

1. 打开 [MongoDB Atlas](https://www.mongodb.com/cloud/atlas) 注册  
2. 创建 **M0 免费集群**  
3. **Database Access** → 添加用户（记住用户名/密码）  
4. **Network Access** → **Add IP Address** → `0.0.0.0/0`（学习用；生产可收紧）  
5. **Connect** → **Drivers** → 复制连接串，形如：  
   `mongodb+srv://user:pass@cluster0.xxx.mongodb.net/?retryWrites=true&w=majority`

### 2. 配置并启动

```bash
cd backend
cp .env.example .env   # 填入 MONGODB_URI、NFT_AUCTION_ADDRESS
export $(grep -v '^#' .env | xargs)   # 或手动 export

go run ./cmd/server
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/auctions
```

### 3. 插入测试数据（可选）

```bash
go run ./cmd/seed
curl http://localhost:8080/api/v1/auctions
```

默认端口 `8080`，CORS 允许 `http://localhost:5173`。

### 4. 链上索引器（阶段 2.3）

在 `backend/.env` 增加 `SEPOLIA_RPC_URL`（与根目录 Infura 地址相同），重启服务：

```bash
go run ./cmd/server
```

日志出现 `indexer enabled` 后，链上新建/出价/结束的拍卖会自动写入 MongoDB。也可手动触发全量同步：

```bash
# 启动时 indexer 会 backfill auctionCount 并轮询新区块
curl http://localhost:8080/api/v1/auctions
```

## 部署合约（需 Node 22+）

```bash
nvm use 22
cp .env.example .env   # 填写 PRIVATE_KEY、ALCHEMY_API_URL
npm run deploy:all
```

## 本地前端

```bash
cd frontend
cp .env.example .env   # 填写 VITE_SEPOLIA_RPC_URL、VITE_API_URL
npm install
npm run dev
```

### 阶段 2.4：前端接 API

1. 先启动后端（见上文），确认 `curl http://localhost:8080/api/v1/auctions` 有数据  
2. 在 `frontend/.env` 设置 `VITE_API_URL=http://localhost:8080`  
3. `npm run dev` 打开页面，列表标题旁显示 **数据来源：后端 API**  
4. 停掉后端或改错 URL 时，会自动 **回退链上 RPC**（并提示 API 不可用）

部署到 Vercel 时增加环境变量 `VITE_API_URL=https://你的后端域名`（后端 `CORS_ORIGINS` 需包含 Vercel 域名）。详见下文 **部署后端到 Render**。

## 部署后端到 Render（免费）

后端跑在 Render 上后，Vercel 前端才能通过公网 URL 访问 API（不能访问你电脑上的 `localhost:8080`）。

### 1. 把代码推到 GitHub

Render 从 GitHub 拉代码部署，确保 `backend/`、`render.yaml` 已 push。

### 2. 创建 Render 服务

**方式 A — Blueprint（推荐）**

1. 打开 [render.com](https://render.com)，用 GitHub 登录  
2. **New → Blueprint** → 选仓库 `web3infoAndDex`  
3. Render 会读取根目录 `render.yaml`，创建名为 `web3infoanddex-api` 的 Web Service  
4. 按提示填写 **Secret** 环境变量（见下表）→ **Apply**

**方式 B — 手动创建 Web Service**

1. **New → Web Service** → 选同一 GitHub 仓库  
2. 设置：

   | 项 | 值 |
   |----|-----|
   | Root Directory | `backend` |
   | Runtime | **Go** |
   | Build Command | `go build -trimpath -ldflags="-s -w" -o server ./cmd/server` |
   | Start Command | `./server` |
   | Instance Type | Free |

3. **Advanced → Health Check Path** 填 `/health`  
4. 在 **Environment** 里添加下表变量 → **Create Web Service**

### 3. Render 环境变量（必填）

在 Render 项目 **Environment** 页面添加（值从本地 `backend/.env` 复制，**不要**把 `.env` 提交 Git）：

| 变量 | 说明 | 示例 |
|------|------|------|
| `MONGODB_URI` | Atlas 连接串 | `mongodb+srv://user:pass@cluster...` |
| `NFT_AUCTION_ADDRESS` | 拍卖合约 | `0x751D5EDA4EFA561702EFfAe3d6096B28206df575` |
| `SEPOLIA_RPC_URL` | Sepolia RPC | 与本地 Infura URL 相同 |
| `CORS_ORIGINS` | 允许的前端域名，逗号分隔 | `https://web3info-and-dex.vercel.app,http://localhost:5173` |

可选：`MONGODB_DB=web3dex`、`SYNC_INTERVAL_SEC=15`、`CHAIN_ID=11155111`（Blueprint 已带默认值）。

`PORT` **不要**自己设 — Render 会自动注入。

### 4. 验证部署

部署完成后会得到公网地址，形如：

`https://web3infoanddex-api.onrender.com`

```bash
curl https://web3infoanddex-api.onrender.com/health
curl https://web3infoanddex-api.onrender.com/api/v1/auctions
```

日志里应出现 `indexer enabled`、`indexer backfill done`。

> **免费版注意**：约 15 分钟无访问会休眠，首次打开要等 ~30 秒唤醒。

### 部署失败排查

在 Render 服务页 → **Logs** 查看具体原因：

| 日志关键词 | 原因 | 处理 |
|------------|------|------|
| `MONGODB_URI is required` | 环境变量未填 | **Environment** 补全 4 个 Secret 后 **Manual Deploy** |
| `mongodb:` / connection refused | Atlas 连不上 | Atlas **Network Access** 加 `0.0.0.0/0`；检查密码 URL 编码 |
| `Killed` / OOM / build timeout | Docker 构建内存不足 | 已改为 **Go 原生运行时**（见最新 `render.yaml`），不要用 Docker |
| Health check failed | 进程启动后马上退出 | 看 Logs 第一行 fatal 错误 |

**若 Blueprint 第一次已失败**：Dashboard 删除失败的 `web3infoanddex-api` → 拉最新代码（含 `runtime: go` 的 `render.yaml`）→ 重新 **New → Blueprint**，**4 个 Secret 必须全部填写**：

```
MONGODB_URI=（从 backend/.env 复制）
NFT_AUCTION_ADDRESS=0x751D5EDA4EFA561702EFfAe3d6096B28206df575
SEPOLIA_RPC_URL=（从 backend/.env 复制）
CORS_ORIGINS=https://你的vercel域名.vercel.app,http://localhost:5173
```

### 5. 让 Vercel 前端连上 Render 后端

1. 打开 [Vercel](https://vercel.com) → 项目 `web3info-and-dex` → **Settings → Environment Variables**  
2. 新增：

   | Name | Value |
   |------|-------|
   | `VITE_API_URL` | `https://web3infoanddex-api.onrender.com`（换成你 Render 实际域名，**不要**末尾 `/`） |

3. **Deployments → Redeploy** 重新构建前端  

4. 打开 Vercel 站点，列表旁应显示 **数据来源：后端 API**

若仍显示链上回退：检查 Render 的 `CORS_ORIGINS` 是否包含你的 Vercel 域名（必须完全一致，含 `https://`）。

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
   | `VITE_API_URL` | **留空**（推荐）— 使用 `frontend/vercel.json` 同源代理 `/api` → Render，避免 CORS；本地开发在 `.env` 填 `http://localhost:8080` |

5. 点 **Deploy**，完成后会得到 `https://xxx.vercel.app`  
6. MetaMask 选 **Sepolia** 后打开该链接即可使用

改环境变量后，在 Vercel 项目里 **Redeploy** 一次才会生效。
