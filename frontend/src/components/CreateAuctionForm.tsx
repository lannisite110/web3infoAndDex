import { useEffect, useState } from "react";
import { parseEther } from "viem";
import {
  useAccount,
  useWriteContract,
  useWaitForTransactionReceipt,
} from "wagmi";
import nftAuctionAbi from "../abi/NFTAuction.json";
import testNftAbi from "../abi/TestNFT.json";
import {
  NFT_AUCTION_ADDRESS,
  TEST_NFT_ADDRESS,
} from "../config/contracts";

type Props = { onSuccess?: () => void };

export function CreateAuctionForm({ onSuccess }: Props) {
  const { address, isConnected } = useAccount();
  const [tokenId, setTokenId] = useState("1");
  const [startPriceEth, setStartPriceEth] = useState("0.01");
  const [durationMin, setDurationMin] = useState("30");
  const [step, setStep] = useState<"idle" | "approve" | "create">("idle");

  const { writeContract, data: txHash, isPending, error, reset } =
    useWriteContract();

  const { isLoading: confirming, isSuccess } = useWaitForTransactionReceipt({
    hash: txHash,
  });

  const nftReady =
    TEST_NFT_ADDRESS !== "0x0000000000000000000000000000000000000000";

  useEffect(() => {
    if (!isSuccess || isPending || confirming) return;

    if (step === "approve") {
      setStep("create");
      reset();
      const durationSec = BigInt(Number(durationMin) * 60);
      writeContract({
        address: NFT_AUCTION_ADDRESS,
        abi: nftAuctionAbi,
        functionName: "createAuction",
        args: [
          TEST_NFT_ADDRESS,
          BigInt(tokenId),
          parseEther(startPriceEth),
          durationSec,
        ],
      });
      return;
    }

    if (step === "create") {
      setStep("idle");
      onSuccess?.();
    }
  }, [
    isSuccess,
    isPending,
    confirming,
    step,
    reset,
    writeContract,
    tokenId,
    startPriceEth,
    durationMin,
    onSuccess,
  ]);

  function handleApproveAndCreate() {
    if (!address || !nftReady) return;
    reset();
    setStep("approve");
    writeContract({
      address: TEST_NFT_ADDRESS,
      abi: testNftAbi,
      functionName: "setApprovalForAll",
      args: [NFT_AUCTION_ADDRESS, true],
    });
  }

  const busy = isPending || confirming;

  if (!isConnected) {
    return (
      <p className="muted">连接钱包后可创建拍卖（需先持有 TestNFT）。</p>
    );
  }

  if (!nftReady) {
    return (
      <p className="muted">
        请先在项目根目录运行 deploy-all，写入 TestNFT 地址。
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
      <button type="button" disabled={busy} onClick={handleApproveAndCreate}>
        {busy
          ? step === "approve"
            ? "授权中…"
            : "创建中…"
          : "授权并创建拍卖"}
      </button>
      {error && (
        <p className="muted" style={{ color: "#f28b82", marginTop: "0.5rem" }}>
          {error.shortMessage ?? error.message}
        </p>
      )}
      <p className="muted" style={{ marginTop: "0.75rem" }}>
        第一次会先发一笔授权交易，再发创建拍卖交易（共 2 笔）。
      </p>
    </div>
  );
}
