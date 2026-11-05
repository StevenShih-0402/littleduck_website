// frontend/components/admin/admin-works.jsx
// 作品管理：列出標記展示的文章、分類篩選、排序調整與移出展示

"use client";

import React from "react";
import { useLake } from "@/store/lake-context";

export default function AdminWorks({ openEditor }) {
  const { db, update } = useLake();
  const wCats = db.settings.catsWorks || [];
  const [kind, setKind] = React.useState("全部");
  const all = db.articles.filter(a => a.showcase);
  const list = kind === "全部" ? all : all.filter(a => a.workCat === kind);

  const unshow = id => update(d => {
    const a = d.articles.find(x => x.id === id);
    if (a) a.showcase = false;
    return d;
  });
  // 重新排序：直接在 articles 陣列中交換（部落格依日期排序，不受影響）
  const move = (id, dir) => update(d => {
    const idxOf = x => d.articles.findIndex(t => t.id === x);
    const shown = d.articles.filter(a => a.showcase && (kind === "全部" || a.workCat === kind));
    const i = shown.findIndex(a => a.id === id);
    const j = i + dir;
    if (i < 0 || j < 0 || j >= shown.length) return d;
    const ai = idxOf(shown[i].id), aj = idxOf(shown[j].id);
    [d.articles[ai], d.articles[aj]] = [d.articles[aj], d.articles[ai]];
    return d;
  });

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
            <div className="flex flex-col gap-0.5">
              <button className="a-dim mono text-[10px] cursor-pointer hover:opacity-70 px-1" onClick={() => move(a.id, -1)}>▲</button>
              <button className="a-dim mono text-[10px] cursor-pointer hover:opacity-70 px-1" onClick={() => move(a.id, 1)}>▼</button>
            </div>
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
      <div className="a-dim mono text-[10px] mt-3">▲▼ 調整前台展示排序 — MAG 網格與 TERM 清單皆依此順序。</div>
    </div>
  );
}
