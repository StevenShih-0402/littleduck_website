// frontend/components/admin/admin-overview.jsx
// 後台總覽：KPI 卡片、14 天流量長條圖、熱門文章與待審留言提示

"use client";

import React from "react";
import { useLake } from "@/store/lake-context";

export default function AdminOverview({ goTab }) {
  const { db } = useLake();
  const days = db.stats.daily;
  const last14 = days.slice(-14);
  const maxV = Math.max(...last14.map(d => d.views));
  const totV = days.reduce((s, d) => s + d.views, 0);
  const totU = days.reduce((s, d) => s + d.visitors, 0);
  const pub = db.articles.filter(a => a.status === "published");
  const top = [...db.articles].sort((a, b) => b.views - a.views).slice(0, 5);
  const pending = db.comments.filter(c => c.status === "pending");

  const Kpi = ({ label, v, sub }) => (
    <div className="a-panel p-4">
      <div className="a-dim mono text-[10px] uppercase tracking-[.15em] mb-1">{label}</div>
      <div className="ntc font-black text-[26px] leading-none">{v}</div>
      <div className="a-dim text-[10.5px] mono mt-1">{sub}</div>
    </div>
  );

  return (
    <div className="space-y-5">
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <Kpi label="30 日瀏覽量" v={totV.toLocaleString()} sub="views / 30d" />
        <Kpi label="30 日訪客" v={totU.toLocaleString()} sub="unique visitors" />
        <Kpi label="已發布文章" v={pub.length} sub={db.articles.length - pub.length + " 篇草稿"} />
        <Kpi label="待審留言" v={pending.length} sub={db.comments.length + " 則留言"} />
      </div>

      <div className="grid grid-cols-12 gap-5">
        {/* traffic bars */}
        <div className="a-panel p-5 col-span-12 lg:col-span-7">
          <div className="flex items-baseline justify-between mb-4">
            <div className="ntc font-black text-[14px]">流量趨勢</div>
            <div className="a-dim mono text-[10px]">最近 14 天 · 假資料示意</div>
          </div>
          <div className="flex items-end gap-1.5 h-32">
            {last14.map((d, i) => (
              <div key={i} className="flex-1 flex flex-col items-center gap-1 group">
                <div className="w-full relative flex items-end" style={{ height: "104px" }}>
                  <div className="w-full transition-all"
                    style={{ height: (d.views / maxV * 100) + "%", background: "var(--a-acc)", opacity: .35 + .65 * (d.views / maxV), borderRadius: "3px 3px 0 0" }}
                    title={d.d + " · " + d.views + " views"}></div>
                </div>
                <div className="a-dim mono text-[8px]">{d.d}</div>
              </div>
            ))}
          </div>
        </div>
        {/* top articles */}
        <div className="a-panel p-5 col-span-12 lg:col-span-5">
          <div className="ntc font-black text-[14px] mb-3">熱門文章</div>
          <div className="space-y-2">
            {top.map((a, i) => (
              <div key={a.id} className="flex items-center gap-2.5">
                <span className="mono text-[11px] a-acc font-bold w-5">{i + 1}</span>
                <span className="ntc text-[12.5px] font-bold flex-1 truncate">{a.title}</span>
                <span className="mono text-[10px] a-dim">{a.views}</span>
              </div>
            ))}
          </div>
          <button className="a-btn w-full mt-4 ntc text-[11px]" onClick={() => goTab("posts")}>管理文章 →</button>
        </div>
      </div>

      {pending.length > 0 && (
        <div className="a-panel p-4 flex items-center gap-3">
          <span className="mono a-acc">◍</span>
          <span className="ntc text-[13px] flex-1">有 <b>{pending.length}</b> 則留言待審核。</span>
          <button className="a-btn a-btn-acc ntc text-[11px]" onClick={() => goTab("cmts")}>前往審核</button>
        </div>
      )}
    </div>
  );
}
