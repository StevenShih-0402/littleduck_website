// frontend/components/front/article-page.jsx
// 文章閱讀頁：橫幅、字級調整、Markdown 內文、閱讀數統計與文章留言區

"use client";

import React from "react";
import { useRouter } from "next/navigation";
import { useLake } from "@/store/lake-context";
import { Lake } from "@/store/lake";
import { SecHead, Markdown } from "@/components/ui";

export default function ArticlePage({ id }) {
  const { db, update, theme } = useLake();
  const router = useRouter();
  const a = db.articles.find(x => x.id === id);
  const night = theme === "night";
  const [fs, setFs] = React.useState(14.5);
  React.useEffect(() => {
    setFs(Number(localStorage.getItem("lake_read_fs")) || 14.5);
  }, []);
  const setFsP = v => { setFs(v); localStorage.setItem("lake_read_fs", v); };
  // count a view once per session
  React.useEffect(() => {
    if (!a) return;
    const k = "lake_viewed_" + a.id;
    if (!sessionStorage.getItem(k)) {
      sessionStorage.setItem(k, "1");
      update(d => {
        const t = d.articles.find(x => x.id === a.id);
        if (t) t.views += 1;
        return d;
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);
  if (!a || a.status !== "published") {
    return (
      <div className="py-16 text-center">
        <div className="mono text-[13px] v-dim mb-4">404 — 找不到這篇文章（可能已下架）。</div>
        <button className="btn ntc" onClick={() => router.push("/blog")}>← 回部落格</button>
      </div>
    );
  }
  return (
    <div className="max-w-[720px] mx-auto">
      <div className="flex items-center justify-between mb-6">
        <button onClick={() => router.push("/blog")} className="mono text-[11px] v-dim hover-rule cursor-pointer tap">
          {night ? "$ cd ../blog" : "← 回部落格目錄"}
        </button>
        <label className="flex items-center gap-1.5 mono text-[10px] v-dim">
          {night ? "--font-size" : "字級"}
          <select className="inp mono text-[11px] cursor-pointer" style={{ width: "auto", padding: "4px 8px" }}
            value={fs} onChange={e => setFsP(Number(e.target.value))}>
            <option value={13}>小</option>
            <option value={14.5}>標準</option>
            <option value={17}>大</option>
            <option value={20}>特大</option>
          </select>
        </label>
      </div>

      {/* Notion-like banner */}
      {a.banner && (
        <div className="relative overflow-hidden rad border v-line mb-6 h-44 sm:h-56">
          <img src={a.banner} alt="" className="w-full h-full object-cover"
            style={a.bannerFade ? { opacity: .45 } : {}} />
          {a.bannerFade && <div className="absolute inset-0" style={{ background: "linear-gradient(to bottom, transparent 40%, var(--bg))" }}></div>}
        </div>
      )}

      {night ? (
        <div className="mono text-[10px] v-dim uppercase tracking-[.2em] mb-3 flex gap-3 flex-wrap">
          <span className="v-acc">$ cat blog/{a.id}.md</span><span>{a.date}</span><span>{a.views} views</span>
        </div>
      ) : (
        <div className="flex items-center gap-3 mb-3">
          <span className="ntc text-[11px] font-black px-2.5 py-1" style={{ background: "var(--ink)", color: "var(--paper)" }}>{a.cat}</span>
          <span className="mono text-[10px] v-dim uppercase tracking-[.2em]">{a.date} · {a.views} VIEWS</span>
        </div>
      )}
      <h1 className="ntc text-[32px] sm:text-[42px] font-black tracking-tight leading-[1.1] mb-3">{a.title}</h1>
      {a.award && <div className="mono text-[11px] v-acc2 mb-4">★ {a.award}</div>}
      <div className="border-t rule-heavy my-6" style={{ borderTopWidth: "var(--heavy)" }}></div>
      <Markdown text={a.body} size={fs} />

      {/* end mark: MAG ■ / TERM EOF */}
      {night ? (
        <div className="mono text-[11px] v-dim mt-8">^D<br /><span className="v-acc">-- EOF --</span><span className="cursor"></span></div>
      ) : (
        <div className="mt-8 flex items-center gap-3">
          <span className="inline-block w-3.5 h-3.5" style={{ background: "var(--ink)" }}></span>
          <span className="mono text-[10px] v-dim uppercase tracking-[.25em]">Fin</span>
        </div>
      )}

      <ArticleComments article={a} />
    </div>
  );
}

// ── per-article comments（預設摺疊，保持閱讀版面乾淨）──
function ArticleComments({ article }) {
  const { db, update, theme } = useLake();
  const night = theme === "night";
  const list = db.comments.filter(c => c.articleId === article.id && c.status === "visible");
  const [open, setOpen] = React.useState(false);
  const [name, setName] = React.useState("");
  const [text, setText] = React.useState("");
  const [sent, setSent] = React.useState(false);

  const submit = () => {
    if (!text.trim()) return;
    update(d => {
      d.comments.unshift({
        id: Lake.uid(), name: name.trim() || "匿名訪客", date: Lake.today(),
        text: text.trim(), status: "pending", reply: "", articleId: article.id,
      });
      return d;
    });
    setName(""); setText(""); setSent(true);
    setTimeout(() => setSent(false), 3500);
  };

  return (
    <div className="border-t rule-heavy mt-10 pt-5" style={{ borderTopWidth: "var(--heavy)" }}>
      <button onClick={() => setOpen(o => !o)}
        className="w-full flex items-center justify-between cursor-pointer tap mb-1 group">
        <span className="mono text-[10.5px] v-dim uppercase tracking-[.2em] group-hover:opacity-75">
          {night ? "$ cat comments/" + article.id + ".log · " + list.length : "讀者留言 · COMMENTS · " + list.length}
        </span>
        <span className="mono text-[11px] v-acc">{open ? (night ? "[-] 摺疊" : "− 摺疊") : (night ? "[+] 展開" : "＋ 展開留言")}</span>
      </button>
      {!open && (
        <div className="mono text-[10.5px] v-dim mb-3">
          {list.length > 0 ? "有 " + list.length + " 則回應。" : "目前還沒有留言，等你搶頭香。"}
        </div>
      )}
      {open && (
      <React.Fragment>
      <div className="space-y-3 mb-5 mt-3">
        {list.map(c => (
          <div key={c.id} className="v-card p-3.5 theme-fade">
            <div className="flex items-baseline gap-2 mb-1">
              <span className="ntc text-[12.5px] font-black">{c.name}</span>
              <span className="mono text-[10px] v-dim">{c.date}</span>
            </div>
            <div className="ntc text-[13px] leading-relaxed">{c.text}</div>
            {c.reply && (
              <div className="mt-2 border-l-2 pl-3" style={{ borderColor: "var(--acc)" }}>
                <span className="mono text-[10px] v-acc mr-2">↳ 屋主</span>
                <span className="ntc text-[12.5px] v-dim">{c.reply}</span>
              </div>
            )}
          </div>
        ))}
        {list.length === 0 && <div className="mono text-[11px] v-dim">{night ? "(empty) — 頭香" : "頭香！"}</div>}
      </div>
      <div className="v-card p-4 theme-fade">
        <div className="grid sm:grid-cols-[180px_1fr] gap-2.5 mb-2.5">
          <input className="inp ntc" placeholder="暱稱（可留空）" value={name} onChange={e => setName(e.target.value)} maxLength={20} />
        </div>
        <textarea className="inp ntc" rows={2} placeholder={night ? "> 留言內容…" : "留言內容…"} value={text} onChange={e => setText(e.target.value)} maxLength={300}></textarea>
        <div className="flex items-center justify-between mt-2.5">
          <div className="mono text-[10px] v-dim">{sent ? "✓ 已送出，待審核後顯示" : text.length + " / 300"}</div>
          <button className="btn btn-acc ntc" onClick={submit}>送出留言</button>
        </div>
      </div>
      </React.Fragment>
      )}
    </div>
  );
}
