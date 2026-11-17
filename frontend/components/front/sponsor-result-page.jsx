// frontend/components/front/sponsor-result-page.jsx
// 贊助付款結果頁（SDP-104）：依 trade_no 查詢付款狀態，顯示感謝／處理中／未完成

"use client";

import React from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useLake } from "@/store/lake-context";
import { HoneyLemonIcon } from "./sponsor-fab";
import { resultView, nextPollDelay } from "@/lib/sponsor.mjs";

export default function SponsorResultPage() {
  const { getPaymentStatus, theme } = useLake();
  const params = useSearchParams();
  const tradeNo = params.get("trade_no") || "";
  // payment：null 表示尚未查到（查詢中或失敗）
  const [payment, setPayment] = React.useState(null);
  const [exhausted, setExhausted] = React.useState(false);
  const [queryFailed, setQueryFailed] = React.useState(false);

  React.useEffect(() => {
    if (!tradeNo) return undefined;
    let cancelled = false;
    let timer = null;
    const run = async (attempt) => {
      let next = null;
      try {
        next = await getPaymentStatus(tradeNo);
        if (cancelled) return;
        setQueryFailed(false);
        setPayment(next);
      } catch {
        if (cancelled) return;
        setQueryFailed(true);
      }
      // 查到 paid／failed／查無即停；pending 或查詢失敗才重查
      if (next && resultView(next.status) !== "pending") return;
      const delay = nextPollDelay(attempt);
      if (delay === null) { setExhausted(true); return; }
      timer = setTimeout(() => run(attempt + 1), delay);
    };
    run(0);
    return () => { cancelled = true; if (timer) clearTimeout(timer); };
  }, [tradeNo, getPaymentStatus]);

  const view = !tradeNo ? "missing" : payment ? resultView(payment.status) : (exhausted ? "unknown" : "loading");

  return (
    <div className="max-w-[480px] mx-auto py-10 text-center theme-fade">
      <div className="v-acc mb-4 flex justify-center"><HoneyLemonIcon size={52} /></div>
      {view === "paid" && (
        <React.Fragment>
          <div className="ntc font-black text-[22px]">感謝你的蜂蜜檸檬！</div>
          <div className="mono text-[12px] v-dim mt-2">
            NT$ {payment.amount} · {theme === "night" ? "payment ok (200)" : "心意已送達"}
          </div>
        </React.Fragment>
      )}
      {(view === "pending" || view === "loading") && (
        <React.Fragment>
          <div className="ntc font-black text-[20px]">付款處理中…</div>
          <div className="ntc text-[13px] v-dim mt-2">正在向綠界確認付款結果，請稍候。</div>
        </React.Fragment>
      )}
      {view === "unknown" && (
        <React.Fragment>
          <div className="ntc font-black text-[20px]">還在確認付款結果</div>
          <div className="ntc text-[13px] v-dim mt-2">
            {queryFailed ? "暫時無法取得付款結果，" : "付款結果尚未回報，"}請稍後重新整理本頁確認。
          </div>
        </React.Fragment>
      )}
      {view === "pending" && exhausted && (
        <div className="ntc text-[13px] v-dim mt-2">付款結果尚未回報，請稍後重新整理本頁確認。</div>
      )}
      {(view === "failed" || view === "missing") && (
        <React.Fragment>
          <div className="ntc font-black text-[20px]">付款未完成</div>
          <div className="ntc text-[13px] v-dim mt-2">這筆贊助沒有完成付款，沒有扣款。</div>
        </React.Fragment>
      )}
      {(view === "paid" || view === "failed" || view === "missing" || exhausted) && (
        <div className="mt-6">
          <Link href="/" className="btn btn-acc ntc inline-block">回首頁</Link>
        </div>
      )}
    </div>
  );
}
