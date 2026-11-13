// frontend/components/admin/admin-features.jsx
// 功能開關管理：列出名稱、說明、key 與狀態，可逐項切換（經 setFeatureEnabled 寫回後端）

"use client";

import React from "react";
import { useLake } from "@/store/lake-context";

export default function AdminFeatures() {
  const { featureList, setFeatureEnabled } = useLake();
  const [busyKey, setBusyKey] = React.useState(null);
  const [err, setErr] = React.useState("");

  const toggle = async (f) => {
    if (busyKey) return;
    setErr(""); setBusyKey(f.key);
    try {
      await setFeatureEnabled(f.key, !f.enabled);
    } catch (e) {
      setErr((e && e.message) || "切換失敗");
    } finally {
      setBusyKey(null);
    }
  };

  return (
    <div>
      <div className="mb-5">
        <div className="ntc font-black text-[20px]">功能開關</div>
        <div className="a-dim text-[12px] ntc mt-1">停用中的功能，管理員在同一分頁的前台仍可預覽。</div>
      </div>
      {err && <div className="text-[11px] mono mb-3" style={{ color: "#d14343" }}>✕ {err}</div>}
      <div className="a-panel divide-y">
        {featureList === null && <div className="p-4 a-dim mono text-[11px]">載入中…</div>}
        {(featureList || []).map(f => {
          const busy = busyKey === f.key;
          return (
            <div key={f.key} className="p-4 flex items-center gap-4">
              <div className="flex-1 min-w-0">
                <div className="flex items-baseline gap-2">
                  <span className="ntc font-black text-[14px]">{f.name}</span>
                  <span className="mono text-[10px] a-dim">{f.key}</span>
                </div>
                <div className="ntc text-[12px] a-dim mt-0.5">{f.description}</div>
              </div>
              <span className="mono text-[11px]" style={f.enabled ? undefined : { opacity: .6 }}>{f.enabled ? "● 啟用中" : "○ 已停用"}</span>
              <button className={"a-btn ntc " + (f.enabled ? "" : "a-btn-acc")} disabled={busy}
                onClick={() => toggle(f)} style={busy ? { opacity: .6, cursor: "not-allowed" } : {}}>
                {busy ? "切換中…" : f.enabled ? "停用" : "啟用"}
              </button>
            </div>
          );
        })}
      </div>
    </div>
  );
}
