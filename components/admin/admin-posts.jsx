// frontend/components/admin/admin-posts.jsx
// 文章管理：狀態篩選、發布/下架切換、刪除與新增文章

"use client";

import React from "react";
import { useLake } from "@/store/lake-context";
import { Lake } from "@/store/lake";
import StatusPill from "./status-pill";

export default function AdminPosts({ openEditor }) {
  const { db, update } = useLake();
  const [filter, setFilter] = React.useState("all");
  const list = db.articles.filter(a => filter === "all" ? true : a.status === filter);

  const toggle = id => update(d => {
    const a = d.articles.find(x => x.id === id);
    if (a) a.status = a.status === "published" ? "draft" : "published";
    return d;
  });
  const del = id => {
    if (!confirm("刪除這篇文章？")) return;
    update(d => { d.articles = d.articles.filter(x => x.id !== id); return d; });
  };
  const create = () => {
    const id = Lake.uid();
    update(d => {
      d.articles.unshift({ id, title: "未命名文章", cat: (d.settings.catsBlog || ["隨筆"])[0], date: Lake.today(), y: 2026, status: "draft", views: 0, blurb: "", award: "", banner: "", bannerFade: false, showcase: false, workCat: "", gallery: [], body: "" });
      return d;
    });
    openEditor(id);
  };

  return (
    <div>
      <div className="flex items-center gap-2 mb-4 flex-wrap">
        <div className="ntc font-black text-[16px] mr-2">文章管理</div>
        {[["all", "全部"], ["published", "已發布"], ["draft", "草稿"]].map(([k, l]) => (
          <button key={k} onClick={() => setFilter(k)}
            className={"mono text-[11px] px-2.5 py-1 cursor-pointer tap " + (filter === k ? "a-acc-bg" : "a-dim border a-line")}
            style={{ borderRadius: "var(--a-rad)" }}>{l}</button>
        ))}
        <div className="flex-1"></div>
        <button className="a-btn a-btn-acc ntc text-[12px]" onClick={create}>＋ 新增文章</button>
      </div>
      <div className="a-panel overflow-hidden">
        {list.map((a, i) => (
          <div key={a.id} className={"flex items-center gap-3 px-4 py-3 " + (i > 0 ? "border-t a-line" : "")}>
            <div className="flex-1 min-w-0">
              <div className="flex items-baseline gap-2 flex-wrap">
                <span className="ntc font-bold text-[13.5px] truncate">{a.title}</span>
                <span className="mono text-[9.5px] a-acc">{a.cat}</span>
                <StatusPill st={a.status} />
              </div>
              <div className="a-dim mono text-[10px] mt-0.5">{a.date} · {a.views} views</div>
            </div>
            <button className="a-btn ntc text-[11px]" onClick={() => openEditor(a.id)}>編輯</button>
            <button className="a-btn ntc text-[11px]" onClick={() => toggle(a.id)}>{a.status === "published" ? "下架" : "發布"}</button>
            <button className="a-btn ntc text-[11px] a-dim" onClick={() => del(a.id)}>刪除</button>
          </div>
        ))}
        {list.length === 0 && <div className="a-dim mono text-[11px] p-6 text-center">沒有符合的文章。</div>}
      </div>
    </div>
  );
}
