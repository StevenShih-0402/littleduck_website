// frontend/components/admin/admin-works.jsx
// 作品管理：列出標記展示的文章、分類篩選與移出展示（改接後端 API）
// 註：前台展示排序需後端排序欄位，本輪 API 尚無，故先移除 ▲▼ 重排控制（後續再補）。

"use client";

import React from "react";
import { useLake } from "@/store/lake-context";

export default function AdminWorks({ openEditor }) {
  const { db, updateArticle } = useLake();
  const wCats = (db.settings && db.settings.catsWorks) || [];
  const [kind, setKind] = React.useState("全部");
  const all = db.articles.filter(a => a.showcase);
  const list = kind === "全部" ? all : all.filter(a => a.workCat === kind);

  const unshow = id => { updateArticle(id, { showcase: false }); };

  return (
    <div>
      <div className="flex items-center gap-2 mb-4 flex-wrap">
        <div className="ntc font-black text-[16px] mr-2">作品管理</div>
        {["全部", ...wCats].map(k => (
          <button key={k} onClick={() => setKind(k)}
            className={"mono text-[11px] px-2.5 py-1 cursor-pointer tap " + (kind === k ? "a-acc-bg" : "a-dim border a-line")}
            style={{ borderRadius: "var(--a-rad)" }}>{k}</button>
        ))}
        <div className="flex-1"></div>
        <div className="a-dim mono text-[10px] hidden sm:block">展示中 {all.filter(a => a.status === "published").length} 篇</div>
      </div>
      <div className="a-panel p-3 mb-4 a-dim ntc text-[12px]">
        作品展示與部落格連動 — 這裡列出所有標記「顯示於作品展示」的文章；封面圖、分類與內文請由「編輯」進入文章編輯器設定。前台點開作品即閱讀全文。
      </div>
      <div className="a-panel overflow-hidden">
        {list.map((a, i) => (
          <div key={a.id} className={"flex items-center gap-3 px-4 py-3 " + (i > 0 ? "border-t a-line" : "") + (a.status === "published" ? "" : " opacity-50")}>
            <div className="flex-1 min-w-0">
              <div className="flex items-baseline gap-2 flex-wrap">
                <span className="ntc font-bold text-[13.5px]">{a.title}</span>
                <span className="mono text-[9.5px] a-acc">{a.workCat}</span>
                {a.status !== "published" && <span className="mono text-[9.5px] a-dim">（草稿 — 前台不顯示）</span>}
              </div>
              <div className="a-dim text-[11px] ntc truncate mt-0.5">{a.blurb}</div>
              <div className="a-dim mono text-[9.5px] mt-0.5">封面 {(a.gallery || []).length} / 10 · {a.views} views</div>
            </div>
            <button className="a-btn ntc text-[11px]" onClick={() => openEditor(a.id)}>編輯</button>
            <button className="a-btn ntc text-[11px]" onClick={() => unshow(a.id)}>移出展示</button>
          </div>
        ))}
        {list.length === 0 && <div className="a-dim mono text-[11px] p-6 text-center">此分類尚無展示作品 — 到文章編輯器勾選「顯示於作品展示」。</div>}
      </div>
    </div>
  );
}
