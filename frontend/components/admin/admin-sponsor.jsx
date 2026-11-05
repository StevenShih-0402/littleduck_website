// frontend/components/admin/admin-sponsor.jsx
// 贊助管理：贊助 KPI、贊助視窗設定（標題/副標/金額按鈕）與贊助紀錄列表

"use client";

import React from "react";
import { useLake } from "@/store/lake-context";

export default function AdminSponsor() {
  const { db, update } = useLake();
  const sp = db.settings.sponsor;
  const dons = db.donations || [];
  const set = (k, v) => update(d => { d.settings.sponsor[k] = v; return d; });
  const setTier = (i, k, v) => update(d => { d.settings.sponsor.tiers[i][k] = v; return d; });

  const total = dons.reduce((s, x) => s + x.amt, 0);
  const month = dons.filter(x => x.date.startsWith("2026 / 07"));
  const monthTotal = month.reduce((s, x) => s + x.amt, 0);
  const avg = dons.length ? Math.round(total / dons.length) : 0;

  const Kpi = ({ label, v, sub }) => (
    <div className="a-panel p-4">
      <div className="a-dim mono text-[10px] uppercase tracking-[.15em] mb-1">{label}</div>
      <div className="ntc font-black text-[26px] leading-none">{v}</div>
      <div className="a-dim text-[10.5px] mono mt-1">{sub}</div>
    </div>
  );

  return (
    <div className="space-y-5">
      <div className="ntc font-black text-[16px]">贊助管理</div>
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <Kpi label="累計贊助" v={"NT$ " + total.toLocaleString()} sub={dons.length + " 筆"} />
        <Kpi label="本月贊助" v={"NT$ " + monthTotal.toLocaleString()} sub={month.length + " 筆 · 7 月"} />
        <Kpi label="平均金額" v={"NT$ " + avg} sub="avg / donation" />
        <Kpi label="最近一筆" v={dons[0] ? "NT$ " + dons[0].amt : "—"} sub={dons[0] ? dons[0].date : "尚無贊助"} />
      </div>

      <div className="grid grid-cols-12 gap-5">
        {/* 設定 */}
        <div className="a-panel p-5 col-span-12 lg:col-span-6 space-y-4">
          <div className="ntc font-black text-[14px]">贊助視窗設定</div>
          <div>
            <label className="block ntc text-[12px] font-bold mb-1">標題</label>
            <input className="a-inp ntc" value={sp.title} onChange={e => set("title", e.target.value)} />
          </div>
          <div>
            <label className="block ntc text-[12px] font-bold mb-1">副標題</label>
            <input className="a-inp ntc" value={sp.subtitle} placeholder="例：人要吃飯才能創作、網站要錢才能運作。" onChange={e => set("subtitle", e.target.value)} />
          </div>
          <div>
            <label className="block ntc text-[12px] font-bold mb-1.5">預設金額按鈕 × 3（金額 + 說明小字）</label>
            <div className="space-y-2">
              {sp.tiers.map((t, i) => (
                <div key={i} className="grid grid-cols-[90px_1fr] gap-2">
                  <input type="number" className="a-inp mono text-[12px]" value={t.amt}
                    onChange={e => setTier(i, "amt", Number(e.target.value) || 0)} />
                  <input className="a-inp ntc" value={t.label} placeholder="例：Claude 訂閱費"
                    onChange={e => setTier(i, "label", e.target.value)} />
                </div>
              ))}
            </div>
            <div className="a-dim mono text-[10px] mt-1.5">前台贊助視窗即時反映。</div>
          </div>
        </div>
        {/* 紀錄 */}
        <div className="a-panel p-5 col-span-12 lg:col-span-6">
          <div className="ntc font-black text-[14px] mb-3">贊助紀錄</div>
          <div className="space-y-2 max-h-[380px] overflow-y-auto">
            {dons.map(x => (
              <div key={x.id} className="flex items-baseline gap-2.5 border-b a-line pb-2">
                <span className="mono text-[10px] a-dim shrink-0">{x.date}</span>
                <span className="ntc text-[12.5px] font-bold flex-1 truncate">{x.name}{x.note ? <span className="a-dim font-normal">「{x.note}」</span> : null}</span>
                <span className="mono text-[12px] a-acc font-bold shrink-0">NT$ {x.amt}</span>
              </div>
            ))}
            {dons.length === 0 && <div className="a-dim mono text-[11px]">尚無贊助紀錄。</div>}
          </div>
        </div>
      </div>
    </div>
  );
}
