// frontend/components/front/home-page.jsx
// 首頁：個人名片 hero、最新動態（清單/網格切換）、分類與留言板預覽

"use client";

import React from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useLake } from "@/store/lake-context";
import { SecHead } from "@/components/ui";

export default function HomePage() {
  const { db, theme } = useLake();
  const router = useRouter();
  const [view, setView] = React.useState("grid");
  const [ready, setReady] = React.useState(false);
  React.useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- 掛載後才讀 localStorage，避免 SSR hydration 不一致
    setView(localStorage.getItem("lake_home_view") || (theme === "night" ? "list" : "grid"));
    setReady(true);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
  React.useEffect(() => { if (ready) localStorage.setItem("lake_home_view", view); }, [view, ready]);
  const pub = db.articles.filter(a => a.status === "published");
  const latest = [...pub].sort((a, b) => b.date < a.date ? -1 : 1).slice(0, 5);
  const cats = {};
  pub.forEach(a => { cats[a.cat] = (cats[a.cat] || 0) + 1; });
  const worksCount = pub.filter(a => a.showcase).length;

  return (
    <div>
      {/* hero: personal card */}
      <div className="grid grid-cols-12 gap-5 sm:gap-7 mb-10">
        <div className="col-span-12 md:col-span-8">
          <div className="mono text-[10.5px] v-dim uppercase tracking-[.25em] mb-3">
            {theme === "night" ? "$ cat lakeside.md --render" : "FEATURE · 小屋門面 / CABIN"}
          </div>
          <h1 className="ntc font-black tracking-tighter leading-[0.95] text-[44px] sm:text-[64px] md:text-[76px]">
            {db.settings.title}
            {theme === "night" && <span className="cursor"></span>}
          </h1>
          <div className="ntc text-[14px] sm:text-[15px] v-dim mt-4 max-w-xl leading-relaxed">
            <span className="mono v-acc">{db.settings.tagline}</span> — {db.settings.subtitle}
          </div>
          <div className="flex flex-wrap gap-2 mt-5">
            <button onClick={() => router.push("/blog")} className="btn btn-acc ntc">開始閱讀 →</button>
            <button onClick={() => router.push("/works")} className="btn ntc">看作品</button>
            <button onClick={() => router.push("/guest")} className="btn ntc">簽到留言</button>
          </div>
        </div>
        {/* profile card — 巴哈小屋味 */}
        <aside className="col-span-12 md:col-span-4">
          <div className="v-card p-5 theme-fade">
            <div className="flex items-center gap-3 mb-4">
              <img src="/assets/duck-logo.jpg" alt="" className="w-14 h-14 rounded-full border-2 v-line" />
              <div>
                <div className="ntc font-black text-[16px]">{db.settings.owner}</div>
                <div className="mono text-[10px] v-dim">a.k.a. {db.settings.aka}</div>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-2 text-center mb-4">
              <div className="border v-soft rad py-2">
                <div className="ntc font-black text-[18px] v-acc">{pub.length}</div>
                <div className="mono text-[9px] v-dim">文章</div>
              </div>
              <div className="border v-soft rad py-2">
                <div className="ntc font-black text-[18px] v-acc">{worksCount}</div>
                <div className="mono text-[9px] v-dim">作品</div>
              </div>
            </div>
            <div className="serif italic text-[14px] v-acc2 leading-snug border-t v-soft pt-3">&quot;{db.settings.motto}&quot;</div>
          </div>
        </aside>
      </div>

      {/* latest + categories */}
      <div className="grid grid-cols-12 gap-5 sm:gap-7">
        <div className="col-span-12 md:col-span-8">
          <div className="relative">
            <SecHead num="01" zh="最新動態" en="" />
            <div className="absolute right-0 top-[14px] flex gap-1">
              {[["list", "≡", "清單"], ["grid", "▦", "網格"]].map(([k, ic, l]) => (
                <button key={k} onClick={() => setView(k)} title={l}
                  className={"mono w-8 h-8 border rad cursor-pointer " + (view === k ? "v-acc-bg border-transparent" : "v-line v-dim hover:opacity-75")}
                  style={{ fontSize: "24px", textAlign: "center", lineHeight: k === "list" ? "27.65px" : "20.05px" }}>
                  {ic}
                </button>
              ))}
            </div>
          </div>
          {view === "list" ? (
            <div className="space-y-0">
              {latest.map(a => (
                <button key={a.id} onClick={() => router.push("/blog/" + a.id)}
                  className="w-full text-left flex items-baseline gap-3 py-3 border-b v-soft cursor-pointer group tap">
                  <span className="mono text-[10px] v-dim shrink-0 w-24">{a.date}</span>
                  <span className="mono text-[9.5px] v-acc shrink-0 border v-line rad px-1.5 py-0.5">{a.cat}</span>
                  <span className="ntc text-[14px] font-bold group-hover:underline flex-1 truncate">{a.title}</span>
                  <span className="mono text-[10px] v-dim shrink-0 hidden sm:block">{a.views} views</span>
                </button>
              ))}
            </div>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {latest.map(a => (
                <button key={a.id} onClick={() => router.push("/blog/" + a.id)}
                  className="v-card p-4 text-left cursor-pointer group theme-fade flex flex-col">
                  <div className="flex items-baseline justify-between gap-2 mb-1.5">
                    <span className="mono text-[9.5px] v-acc border v-line rad px-1.5 py-0.5">{a.cat}</span>
                    <span className="mono text-[10px] v-dim">{a.date}</span>
                  </div>
                  <div className="ntc text-[15px] font-black group-hover:underline leading-snug">{a.title}</div>
                  <div className="ntc text-[11.5px] v-dim leading-relaxed mt-1.5 line-clamp-2 flex-1">{a.blurb}</div>
                  <div className="mono text-[9.5px] v-dim mt-2 pt-2 border-t v-soft">{a.views} views</div>
                </button>
              ))}
            </div>
          )}
          <Link href="/blog" className="mono text-[11px] v-acc hover-rule mt-4 cursor-pointer inline-block">全部文章 →</Link>
        </div>
        <aside className="col-span-12 md:col-span-4">
          <SecHead num="02" zh="分類" en="CATEGORIES" />
          <div className="space-y-1.5">
            {Object.entries(cats).map(([c, n]) => (
              <button key={c} onClick={() => router.push("/blog")}
                className="w-full flex justify-between items-center py-2 px-3 v-card cursor-pointer hover:opacity-80 tap theme-fade">
                <span className="ntc text-[13px] font-bold">{c}</span>
                <span className="mono text-[11px] v-dim">{n} 篇</span>
              </button>
            ))}
          </div>
          <div className="mono text-[10px] v-dim uppercase tracking-[.2em] mt-7 mb-2">— Guestbook —</div>
          <div className="v-card p-4 theme-fade">
            {db.comments.filter(c => c.status === "visible").slice(0, 1).map(c => (
              <div key={c.id}>
                <div className="ntc text-[12.5px] leading-relaxed">「{c.text}」</div>
                <div className="mono text-[10px] v-dim mt-2">— {c.name} · {c.date}</div>
              </div>
            ))}
            <Link href="/guest" className="mono text-[11px] v-acc hover-rule mt-3 cursor-pointer inline-block">去留言板 →</Link>
          </div>
        </aside>
      </div>
    </div>
  );
}
