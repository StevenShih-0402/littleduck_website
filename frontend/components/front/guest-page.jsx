// frontend/components/front/guest-page.jsx
// 訪客留言板：投書表單（送出後待審核）與已通過留言列表

"use client";

import React from "react";
import { useLake } from "@/store/lake-context";
import { Lake } from "@/store/lake";
import { SecHead } from "@/components/ui";

export default function GuestPage() {
  const { db, postComment, theme } = useLake();
  const [name, setName] = React.useState("");
  const [text, setText] = React.useState("");
  const [sent, setSent] = React.useState(false);
  const [err, setErr] = React.useState("");
  // 留言板留言：articleId 為空字串/null 皆視為留言板
  const visible = db.comments.filter(c => c.status === "visible" && !c.articleId);

  const submit = async () => {
    if (!text.trim()) return;
    setErr("");
    try {
      await postComment({ name: name.trim() || "匿名訪客", text: text.trim(), articleId: "" });
      setName(""); setText(""); setSent(true);
      setTimeout(() => setSent(false), 3500);
    } catch (e) {
      setErr(e.message || "送出失敗，請稍後再試。");
    }
  };

  return (
    <div className="max-w-[760px]">
      <SecHead num="04" zh="訪客留言板" en="GUESTBOOK"
        sub={theme === "night" ? "$ sign --guestbook — 留言送出後由屋主審核再顯示。" : "讀者來函 — 路過簽到、感想、敲碗都可以，審核後刊登。"} />
      {/* composer */}
      <div className="v-card p-4 sm:p-5 mb-8 theme-fade">
        <div className="mono text-[10.5px] v-dim mb-3">{theme === "night" ? "$ cat >> guestbook.log" : "LETTERS · 投書欄"}</div>
        <div className="grid sm:grid-cols-[200px_1fr] gap-3 mb-3">
          <input className="inp ntc" placeholder="暱稱（可留空）" value={name} onChange={e => setName(e.target.value)} maxLength={20} />
          <div className="mono text-[10px] v-dim self-center hidden sm:block">{Lake.today()} · 訪客</div>
        </div>
        <textarea className="inp ntc" rows={3} placeholder="說點什麼吧…" value={text} onChange={e => setText(e.target.value)} maxLength={300}></textarea>
        <div className="flex items-center justify-between mt-3">
          <div className="mono text-[10px] v-dim">{text.length} / 300</div>
          <button className="btn btn-acc ntc" onClick={submit}>送出留言</button>
        </div>
        {sent && <div className="mono text-[11px] v-acc mt-2">✓ 已送出，待屋主審核後顯示。</div>}
        {err && <div className="mono text-[11px] mt-2" style={{ color: "#d14343" }}>✕ {err}</div>}
      </div>
      {/* messages */}
      <div className="space-y-4">
        {visible.map(c => (
          <div key={c.id} className="v-card p-4 theme-fade">
            <div className="flex items-baseline gap-3 mb-1.5">
              <div className="w-7 h-7 rounded-full v-acc-bg flex items-center justify-center ntc text-[12px] font-black shrink-0 self-center">
                {c.name.slice(0, 1)}
              </div>
              <span className="ntc text-[13px] font-black">{c.name}</span>
              <span className="mono text-[10px] v-dim">{c.date}</span>
            </div>
            <div className="ntc text-[13.5px] leading-relaxed pl-10">{c.text}</div>
            {c.reply && (
              <div className="mt-3 ml-10 border-l-2 pl-3" style={{ borderColor: "var(--acc)" }}>
                <div className="mono text-[10px] v-acc mb-0.5">↳ 屋主回覆</div>
                <div className="ntc text-[13px] v-dim leading-relaxed">{c.reply}</div>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
