// frontend/components/admin/admin-shell.jsx
// 後台外框：登入頁、三種風格 skin 切換、側欄導覽與各管理分頁的掛載

"use client";

import React from "react";
import { useRouter } from "next/navigation";
import { useLake } from "@/store/lake-context";
import AdminOverview from "./admin-overview";
import AdminPosts from "./admin-posts";
import AdminEditor from "./admin-editor";
import AdminWorks from "./admin-works";
import AdminComments from "./admin-comments";
import AdminSponsor from "./admin-sponsor";
import AdminSettings from "./admin-settings";
import AdminFeatures from "./admin-features";

function AdminLogin() {
  const router = useRouter();
  const { login } = useLake();
  const [username, setUsername] = React.useState("");
  const [pw, setPw] = React.useState("");
  const [err, setErr] = React.useState("");
  const [busy, setBusy] = React.useState(false);

  const submit = async () => {
    if (busy) return;
    setErr("");
    setBusy(true);
    try {
      await login(username.trim(), pw);
      // 成功後 authed 由 context 更新，AdminShell 會改渲染主畫面
    } catch (e) {
      setErr(e.message || "登入失敗");
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="adm adm-neutral flex items-center justify-center min-h-screen px-4">
      <div className="a-panel p-8 w-full max-w-[380px]">
        <div className="flex items-center gap-3 mb-6">
          <img src="/assets/duck-logo.jpg" alt="" className="w-10 h-10 rounded-full" />
          <div>
            <div className="ntc font-black text-[16px]">湖畔後台</div>
            <div className="a-dim text-[11px] mono">lakeside · admin console</div>
          </div>
        </div>
        <label className="block text-[12px] font-bold mb-1.5 ntc">管理員帳號</label>
        <input type="text" className="a-inp mb-3" value={username} placeholder="username" autoFocus
          onChange={e => setUsername(e.target.value)}
          onKeyDown={e => e.key === "Enter" && submit()} />
        <label className="block text-[12px] font-bold mb-1.5 ntc">管理員密碼</label>
        <input type="password" className="a-inp mb-2" value={pw} placeholder="••••••••"
          onChange={e => setPw(e.target.value)}
          onKeyDown={e => e.key === "Enter" && submit()} />
        {err && <div className="text-[11px] mono mb-3" style={{ color: "#d14343" }}>✕ {err}</div>}
        <button className="a-btn a-btn-acc w-full ntc" onClick={submit} disabled={busy}
          style={busy ? { opacity: .6, cursor: "not-allowed" } : {}}>
          {busy ? "登入中…" : "登入後台 →"}
        </button>
        <button className="a-btn w-full mt-2 ntc" onClick={() => router.push("/")}>← 回前台</button>
      </div>
    </div>
  );
}

const ADMIN_TABS = [
  { key: "over", zh: "總覽", icon: "◧" },
  { key: "posts", zh: "文章管理", icon: "≡" },
  { key: "works", zh: "作品管理", icon: "▦" },
  { key: "cmts", zh: "留言管理", icon: "◍" },
  { key: "spon", zh: "贊助管理", icon: "◈" },
  { key: "set", zh: "網站設定", icon: "⚙" },
  { key: "flags", zh: "功能開關", icon: "⚑" },
];

export default function AdminShell() {
  const { db, authed, logout, reloadAdmin } = useLake();
  const router = useRouter();
  const [tab, setTab] = React.useState("over");
  const [skin, setSkin] = React.useState("neutral");
  const [editId, setEditId] = React.useState(null); // article being edited
  React.useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- 掛載後才讀 localStorage，避免 SSR hydration 不一致
    setSkin(localStorage.getItem("lake_admin_skin") || "neutral");
  }, []);
  // 進入後台時（已登入）確保後台資料（草稿、全部留言、贊助、真實統計）已載入
  React.useEffect(() => {
    if (authed) reloadAdmin();
  }, [authed, reloadAdmin]);
  const setSkinP = s => { setSkin(s); localStorage.setItem("lake_admin_skin", s); };

  if (!authed) return <AdminLogin />;

  const pendingC = db.comments.filter(c => c.status === "pending").length;
  const openEditor = id => { setEditId(id); setTab("editor"); };

  return (
    <div className={"adm adm-" + skin}>
      {/* topbar */}
      <div className="a-panel rounded-none border-x-0 border-t-0 sticky top-0 z-40">
        <div className="max-w-[1200px] mx-auto px-4 sm:px-6 min-h-14 py-2 flex flex-wrap items-center gap-x-3 gap-y-2">
          <img src="/assets/duck-logo.jpg" alt="" className="w-8 h-8 rounded-full" />
          <div className="ntc font-black text-[14px]">湖畔後台</div>
          <div className="a-dim mono text-[10px] hidden sm:block">admin console</div>
          <div className="flex-1"></div>
          {/* skin explorer */}
          <div className="flex items-center gap-1 mono text-[10px] mr-2">
            <span className="a-dim hidden sm:inline mr-1">風格</span>
            {[["neutral", "中性"], ["day", "MAG"], ["night", "TERM"]].map(([k, l]) => (
              <button key={k} onClick={() => setSkinP(k)}
                className={"px-2 py-1 cursor-pointer tap " + (skin === k ? "a-acc-bg rounded" : "a-dim hover:opacity-75")}
                style={skin === k ? { borderRadius: "var(--a-rad)" } : {}}>
                {l}
              </button>
            ))}
          </div>
          <button className="a-btn ntc text-[11px]" onClick={() => router.push("/")}>看前台</button>
          <button className="a-btn ntc text-[11px]" onClick={() => logout()}>登出</button>
        </div>
      </div>

      <div className="max-w-[1200px] mx-auto px-4 sm:px-6 py-6 grid grid-cols-12 gap-5">
        {/* sidebar */}
        <nav className="col-span-12 lg:col-span-2 flex lg:flex-col gap-1 overflow-x-auto no-scrollbar">
          {ADMIN_TABS.map(t => (
            <button key={t.key} onClick={() => { setTab(t.key); setEditId(null); }}
              className={"tap text-left px-3 py-2 ntc text-[13px] font-bold whitespace-nowrap cursor-pointer flex items-center gap-2 " +
                (tab === t.key || (tab === "editor" && t.key === "posts") ? "a-acc-bg" : "a-dim hover:opacity-75")}
              style={{ borderRadius: "var(--a-rad)" }}>
              <span className="mono">{t.icon}</span>{t.zh}
              {t.key === "cmts" && pendingC > 0 && (
                <span className="mono text-[9px] px-1.5 rounded-full" style={{ background: "var(--a-acc)", color: "var(--a-acc-c)", opacity: tab === "cmts" ? .7 : 1 }}>{pendingC}</span>
              )}
            </button>
          ))}
        </nav>
        {/* content */}
        <div className="col-span-12 lg:col-span-10">
          {tab === "over" && <AdminOverview goTab={setTab} />}
          {tab === "posts" && <AdminPosts openEditor={openEditor} />}
          {tab === "editor" && <AdminEditor id={editId} back={() => setTab("posts")} />}
          {tab === "works" && <AdminWorks openEditor={openEditor} />}
          {tab === "cmts" && <AdminComments />}
          {tab === "spon" && <AdminSponsor />}
          {tab === "set" && <AdminSettings />}
          {tab === "flags" && <AdminFeatures />}
        </div>
      </div>
    </div>
  );
}
