// frontend/components/admin/admin-comments.jsx
// 留言管理：待審核/已顯示/已隱藏分組、通過/隱藏/刪除與屋主回覆

"use client";

import React from "react";
import { useLake } from "@/store/lake-context";

export default function AdminComments() {
  const { db, update } = useLake();
  const [replying, setReplying] = React.useState(null);
  const [replyText, setReplyText] = React.useState("");
  const artName = id => {
    if (!id) return null;
    const a = db.articles.find(x => x.id === id);
    return a ? a.title : "(已刪文章)";
  };
  const setStatus = (id, st) => update(d => {
    const c = d.comments.find(x => x.id === id);
    if (c) c.status = st;
    return d;
  });
  const del = id => {
    if (!confirm("刪除這則留言？")) return;
    update(d => { d.comments = d.comments.filter(x => x.id !== id); return d; });
  };
  const saveReply = id => {
    update(d => { const c = d.comments.find(x => x.id === id); if (c) c.reply = replyText.trim(); return d; });
    setReplying(null); setReplyText("");
  };
  const groups = [
    ["pending", "待審核"],
    ["visible", "已顯示"],
    ["hidden", "已隱藏"],
  ];

  return (
    <div className="space-y-6">
      <div className="ntc font-black text-[16px]">留言管理</div>
      {groups.map(([st, label]) => {
        const list = db.comments.filter(c => c.status === st);
        if (list.length === 0) return null;
        return (
          <div key={st}>
            <div className="a-dim mono text-[10.5px] uppercase tracking-[.2em] mb-2">{label} · {list.length}</div>
            <div className="a-panel overflow-hidden">
              {list.map((c, i) => (
                <div key={c.id} className={"px-4 py-3 " + (i > 0 ? "border-t a-line" : "")}>
                  <div className="flex items-center gap-3">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-baseline gap-2 flex-wrap">
                        <span className="ntc font-bold text-[13px]">{c.name}</span>
                        <span className="a-dim mono text-[10px]">{c.date}</span>
                        <span className="mono text-[9px] px-1.5 py-0.5 border a-line a-dim" style={{ borderRadius: "var(--a-rad)" }}>
                          {artName(c.articleId) ? "文章：" + artName(c.articleId) : "留言板"}
                        </span>
                      </div>
                      <div className="ntc text-[12.5px] mt-0.5">{c.text}</div>
                      {c.reply && <div className="a-acc text-[11.5px] ntc mt-1">↳ {c.reply}</div>}
                    </div>
                    {st !== "visible" && <button className="a-btn a-btn-acc ntc text-[11px]" onClick={() => setStatus(c.id, "visible")}>通過</button>}
                    {st === "visible" && (
                      <button className="a-btn ntc text-[11px]" onClick={() => { setReplying(c.id); setReplyText(c.reply || ""); }}>回覆</button>
                    )}
                    {st !== "hidden" && <button className="a-btn ntc text-[11px]" onClick={() => setStatus(c.id, "hidden")}>隱藏</button>}
                    <button className="a-btn ntc text-[11px] a-dim" onClick={() => del(c.id)}>刪除</button>
                  </div>
                  {replying === c.id && (
                    <div className="mt-3 flex gap-2">
                      <input className="a-inp ntc" value={replyText} placeholder="屋主回覆…" autoFocus
                        onChange={e => setReplyText(e.target.value)}
                        onKeyDown={e => e.key === "Enter" && saveReply(c.id)} />
                      <button className="a-btn a-btn-acc ntc text-[11px]" onClick={() => saveReply(c.id)}>送出</button>
                      <button className="a-btn ntc text-[11px]" onClick={() => setReplying(null)}>取消</button>
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        );
      })}
    </div>
  );
}
