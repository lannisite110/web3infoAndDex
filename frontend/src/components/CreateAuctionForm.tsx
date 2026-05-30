import { useEffect, useRef, useState } from "react";
import { parseEther } from "viem";
import {
  useAccount,
  useChainId,
  useReadContract,
  useSimulateContract,
  useWriteContract,
  useWaitForTransactionReceipt,
} from "wagmi";
import nftAuctionAbi from "../abi/NFTAuction.json";
import testNftAbi from "../abi/TestNFT.json";
import {
  NFT_AUCTION_ADDRESS,
  SEPOLIA_CHAIN_ID,
  TEST_NFT_ADDRESS,
} from "../config/contracts";
import { parseTokenId } from "../utils/tokenId";

type Props = { onSuccess?: () => void };

export function CreateAuctionForm({ onSuccess }: Props) {
  const { address, isConnected } = useAccount();
  const chainId = useChainId();
  const onSepolia = chainId === SEPOLIA_CHAIN_ID;

  const [tokenId, setTokenId] = useState("1");
  const [startPriceEth, setStartPriceEth] = useState("0.01");
  const [durationMin, setDurationMin] = useState("30");

  const parsedTokenId = parseTokenId(tokenId);
  const onSuccessRef = useRef(onSuccess);
  onSuccessRef.current = onSuccess;

  const lastActionRef = useRef<"approve" | "create" | null>(null);
  const handledTxRef = useRef<string | null>(null);

  const { writeContract, data: txHash, isPending, error, reset } =
    useWriteContract();

  const { isLoading: confirming, isSuccess } = useWaitForTransactionReceipt({
    hash: txHash,
  });

  const { data: isApproved, refetch: refetchApproval } = useReadContract({
    address: TEST_NFT_ADDRESS,
    abi: testNftAbi,
    functionName: "isApprovedForAll",
    args: address ? [address, NFT_AUCTION_ADDRESS] : undefined,
    query: { enabled: Boolean(address && onSepolia) },
  });

  const {
    data: ownerOf,
    isError: ownerOfFailed,
    error: ownerOfErr,
    isLoading: ownerOfLoading,
  } = useReadContract({
    address: TEST_NFT_ADDRESS,
    abi: testNftAbi,
    functionName: "ownerOf",
    args: parsedTokenId !== null ? [parsedTokenId] : undefined,
    query: { enabled: onSepolia && parsedTokenId !== null },
  });

  const ownsToken =
    !ownerOfFailed &&
    ownerOf &&
    address &&
    (ownerOf as string).toLowerCase() === address.toLowerCase();

  const rpcLikelyBroken =
    ownerOfFailed &&
    (ownerOfErr?.message?.includes("Failed to fetch") ||
      ownerOfErr?.message?.includes("HTTP request failed"));

  const approveSim = useSimulateContract({
    address: TEST_NFT_ADDRESS,
    abi: testNftAbi,
    functionName: "setApprovalForAll",
    args: [NFT_AUCTION_ADDRESS, true],
    query: { enabled: isConnected && onSepolia && !isApproved },
  });

  const durationSec = BigInt(Math.max(1, Number(durationMin) || 1) * 60);
  const createSim = useSimulateContract({
    address: NFT_AUCTION_ADDRESS,
    abi: nftAuctionAbi,
    functionName: "createAuction",
    args:
      parsedTokenId !== null
        ? [
            TEST_NFT_ADDRESS,
            parsedTokenId,
            parseEther(startPriceEth || "0"),
            durationSec,
          ]
        : undefined,
    query: {
      enabled:
        isConnected &&
        onSepolia &&
        Boolean(isApproved) &&
        Boolean(ownsToken) &&
        parsedTokenId !== null,
    },
  });

  const nftReady =
    TEST_NFT_ADDRESS !== "0x0000000000000000000000000000000000000000";

  const busy = isPending || confirming;

  function runApprove() {
    if (!approveSim.data?.request) return;
    lastActionRef.current = "approve";
    reset();
    writeContract(approveSim.data.request);
  }

  function runCreate() {
    if (!createSim.data?.request) return;
    lastActionRef.current = "create";
    reset();
    writeContract(createSim.data.request);
  }

  useEffect(() => {
    if (!isSuccess || !txHash) return;
    if (handledTxRef.current === txHash) return;
    handledTxRef.current = txHash;

    void refetchApproval();

    if (lastActionRef.current === "create") {
      onSuccessRef.current?.();
    }
    lastActionRef.current = null;
    reset();
  }, [isSuccess, txHash, refetchApproval, reset]);

  if (!isConnected) {
    return (
      <p className="muted">连接钱包后可创建拍卖（需先持有 TestNFT）。</p>
    );
  }

  if (!onSepolia) {
    return (
      <p className="muted">
        请先在 MetaMask 切换到 <strong>Sepolia</strong>，再操作。
      </p>
    );
  }

  if (!nftReady) {
    return (
      <p className="muted">
        请配置 TestNFT 合约地址（deploy-all 或 Vercel 环境变量）。
      </p>
    );
  }

  return (
    <div>
      <div className="field-grid">
        <div>
          <label htmlFor="tokenId">NFT Token ID</label>
          <input
            id="tokenId"
            value={tokenId}
            onChange={(e) => setTokenId(e.target.value)}
          />
        </div>
        <div>
          <label htmlFor="price">起拍价 (ETH)</label>
          <input
            id="price"
            value={startPriceEth}
            onChange={(e) => setStartPriceEth(e.target.value)}
          />
        </div>
      </div>
      <label htmlFor="duration">拍卖时长 (分钟)</label>
      <input
        id="duration"
        value={durationMin}
        onChange={(e) => setDurationMin(e.target.value)}
      />

      {tokenId && parsedTokenId === null && (
        <p className="muted" style={{ color: "#f28b82" }}>
          Token ID 请填写数字，例如 1、2、3。
        </p>
      )}

      {rpcLikelyBroken && (
        <p className="muted" style={{ color: "#f28b82" }}>
          无法连接 Sepolia RPC。Vercel 请配置环境变量{" "}
          <code>VITE_SEPOLIA_RPC_URL</code> 后 Redeploy（不要用公共
          rpc.sepolia.org）。
        </p>
      )}

      {parsedTokenId !== null &&
        !ownerOfLoading &&
        !rpcLikelyBroken &&
        !ownsToken && (
          <p className="muted" style={{ color: "#f28b82" }}>
            当前钱包不拥有 Token #{tokenId}。在项目根目录执行：
            <br />
            <code>
              TEST_NFT_ADDRESS={TEST_NFT_ADDRESS} npm run mint:nfts
            </code>
          </p>
        )}

      {isSuccess && confirming === false && (
        <p className="muted" style={{ color: "#8ab4f8" }}>
          上一笔交易已上链，可继续操作或到下方列表查看。
        </p>
      )}

      <p className="muted">
        授权状态：{isApproved ? "已授权 ✓" : "未授权（先点第 1 步）"}
      </p>

      <div className="row">
        <button
          type="button"
          disabled={busy || isApproved || !approveSim.data?.request}
          onClick={runApprove}
        >
          {busy ? "提交中…" : "1. 授权拍卖合约"}
        </button>
        <button
          type="button"
          disabled={
            busy || !isApproved || !ownsToken || !createSim.data?.request
          }
          onClick={runCreate}
        >
          {busy ? "提交中…" : "2. 创建拍卖"}
        </button>
      </div>

      {approveSim.error && !isApproved && (
        <p className="muted" style={{ color: "#f28b82" }}>
          授权模拟失败：{approveSim.error.message}
        </p>
      )}
      {createSim.error && isApproved && (
        <p className="muted" style={{ color: "#f28b82" }}>
          创建模拟失败：{createSim.error.message}
        </p>
      )}
      {error && (
        <p className="muted" style={{ color: "#f28b82" }}>
          {error.shortMessage ?? error.message}
        </p>
      )}

      <p className="muted" style={{ marginTop: "0.75rem" }}>
        第 1 笔确认上链后，再点第 2 步。若整页变黑，请刷新；已修复自动刷新的
        bug。
      </p>
    </div>
  );
}
