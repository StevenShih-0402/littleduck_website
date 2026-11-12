// frontend/components/front/front-shell.jsx
// 前台外框：公告列、頁首導覽（含手機抽屜）、主題切換、頁尾與贊助按鈕

"use client";

import React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useLake } from "@/store/lake-context";
import { SITE } from "@/data/site";
import { classifyHref } from "@/lib/safe-url.mjs";
import SponsorFab from "./sponsor-fab";

const NAV = [
  { key: "home", href: "/", zh: "首頁", en: "HOME" },
  { key: "blog", href: "/blog", zh: "部落格", en: "BLOG" },
  { key: "works", href: "/works", zh: "作品展示", en: "WORKS" },
  { key: "guest", href: "/guest", zh: "留言板", en: "GUESTBOOK" },
];

function activeKey(pathname) {
  if (pathname.startsWith("/blog")) return "blog";
  if (pathname.startsWith("/works")) return "works";
  if (pathname.startsWith("/guest")) return "guest";
  return "home";
}

export default function FrontShell({ children }) {
  const { db, theme, setTheme, loading, error, reportPageview } = useLake();
  const [drawer, setDrawer] = React.useState(false);
  const pathname = usePathname() || "/";
  const active = activeKey(pathname);

  // 前台導覽時回報瀏覽（fire-and-forget，不影響渲染）
  React.useEffect(() => {
    reportPageview(pathname);
  }, [pathname, reportPageview]);

  const settings = db.settings;

  // 載入中 / 錯誤的全頁狀態（避免 null settings 造成崩潰）
  if (loading && !settings) {
    return (
      <div className="min-h-screen v-bg theme-fade flex items-center justify-center">
        <div className="mono text-[12px] v-dim">載入中…</div>
      </div>
    );
  }
  if (error && !settings) {
    return (
      <div className="min-h-screen v-bg theme-fade flex items-center justify-center px-4">
        <div className="text-center">
          <div className="mono text-[12px] v-dim mb-2">無法載入網站內容</div>
          <div className="ntc text-[13px] v-dim">{error}</div>
        </div>
      </div>
    );
  }
  if (!settings) return null;

  return (
    <div className="min-h-screen v-bg theme-fade">
      {/* announcement */}
      {db.settings.announcement && (
        <div className="v-acc-bg mono text-[11px] text-center py-1.5 px-4 tracking-wide">
          {theme === "night" ? "[notice] " + db.settings.announcement : "◈ " + db.settings.announcement}
        </div>
      )}
      {/* header */}
      <header className="border-b rule-heavy v-paper theme-fade sticky top-0 z-40" style={{ borderBottomWidth: "var(--heavy)" }}>
        <div className="max-w-[1100px] mx-auto px-4 sm:px-6 h-16 flex items-center gap-3">
          <button className="lg:hidden tap px-2 v-acc mono text-[18px]" onClick={() => setDrawer(d => !d)} aria-label="menu">☰</button>
          <img src="/assets/duck-logo.jpg" alt="" className="w-9 h-9 rounded-full border v-line" />
          <Link href="/" className="leading-tight mr-4 cursor-pointer block">
            <div className="ntc font-black text-[15px] tracking-tight">{db.settings.title}</div>
            <div className="mono text-[9.5px] v-dim uppercase tracking-[.18em] hidden sm:block">{db.settings.titleEn}</div>
          </Link>
          <nav className="hidden lg:flex items-center gap-1 flex-1">
            {NAV.map(n => (
              <Link key={n.key} href={n.href}
                className={"tap px-3 ntc text-[13px] font-bold cursor-pointer border-b-2 whitespace-nowrap flex items-center " + (active === n.key ? "v-acc" : "v-dim hover:opacity-75")}
                style={{ borderColor: active === n.key ? "var(--acc)" : "transparent" }}>
                {n.zh}<span className="mono text-[9px] ml-1.5 opacity-60">{n.en}</span>
              </Link>
            ))}
          </nav>
          <div className="flex-1 lg:flex-none"></div>
          {/* theme toggle: day = MAG · night = TERM */}
          <button onClick={() => setTheme(theme === "night" ? "day" : "night")}
            className="tap mono text-[11px] px-3 border v-line rad cursor-pointer flex items-center gap-2">
            <span className={theme === "day" ? "v-acc font-bold" : "v-dim"}>MAG</span>
            <span className="v-dim">/</span>
            <span className={theme === "night" ? "v-acc font-bold" : "v-dim"}>TERM</span>
          </button>
          <Link href="/admin" className="tap mono text-[11px] px-3 v-dim hover:opacity-75 cursor-pointer hidden sm:flex items-center">後台 →</Link>
        </div>
        {drawer && (
          <div className="lg:hidden border-t v-soft v-paper px-4 py-2 flex flex-col">
            {NAV.map(n => (
              <Link key={n.key} href={n.href} onClick={() => setDrawer(false)}
                className={"tap text-left ntc text-[14px] font-bold py-2 " + (active === n.key ? "v-acc" : "v-dim")}>
                {n.zh}
              </Link>
            ))}
            <Link href="/admin" onClick={() => setDrawer(false)} className="tap text-left mono text-[12px] v-dim py-2">後台 →</Link>
          </div>
        )}
      </header>

      <main className="max-w-[1100px] mx-auto px-4 sm:px-6 py-8 sm:py-10">{children}</main>

      <SponsorFab />

      <footer className="border-t rule-heavy mt-6" style={{ borderTopWidth: "var(--heavy)" }}>
        <div className="max-w-[1100px] mx-auto px-4 sm:px-6 py-5 flex flex-col sm:flex-row gap-2 justify-between mono text-[10.5px] uppercase tracking-[.15em] v-dim">
          <div>{SITE.copyright}</div>
          <div className="flex gap-4">
            {/* SDP-105：刻意偏離設計稿（原本 preventDefault）。依 classifyHref 渲染，佔位或不安全網址顯示為不可點的文字 */}
            {(db.settings.social || []).map(s => {
              const { kind, href } = classifyHref(s.url);
              if (kind === "internal") return <Link key={s.label} href={href} className="hover-rule cursor-pointer">{s.label}</Link>;
              if (kind === "external") return <a key={s.label} href={href} target="_blank" rel="noopener noreferrer" className="hover-rule cursor-pointer">{s.label}</a>;
              return <span key={s.label}>{s.label}</span>;
            })}
          </div>
        </div>
      </footer>
    </div>
  );
}
