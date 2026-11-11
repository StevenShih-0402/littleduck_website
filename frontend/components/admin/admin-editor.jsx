// frontend/components/admin/admin-editor.jsx
// 文章編輯器：中繼資料表單、橫幅設定、作品展示連動、Markdown 編輯與即時預覽（含全螢幕）
// 資料改接後端：以本機 draft 狀態維持輸入即時反應，變更以 debounce 經 updateArticle 寫回。

"use client";

import React from "react";
import { useLake } from "@/store/lake-context";
import { Markdown } from "@/components/ui";
import StatusPill from "./status-pill";

const SAVE_DELAY = 700; // ms：輸入停止後才寫回後端，避免每次按鍵都打 API

export default function AdminEditor({ id, back }) {
  const { db, updateArticle } = useLake();
  const stored = db.articles.find(x => x.id === id);

  // 本機 draft：輸入即時反映在表單與預覽，不等 API 往返
  const [draft, setDraft] = React.useState(stored || null);
  const [saving, setSaving] = React.useState(false);
  const timer = React.useRef(null);
  const pending = React.useRef(null);

  // 切換文章時同步 draft（以 id 為界，避免把別篇的輸入帶過去）：
  // 在 render 期間比對前一次的 id，取代「effect 內 setState」
  const [prevId, setPrevId] = React.useState(id);
  if (id !== prevId) {
    setPrevId(id);
    setDraft(stored || null);
  }

  const [showPrev, setShowPrev] = React.useState(true);
  const [full, setFull] = React.useState(false);
  const [edFs, setEdFs] = React.useState(12.5);
  React.useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- 掛載後才讀 localStorage，避免 SSR hydration 不一致
    setEdFs(Number(localStorage.getItem("lake_editor_fs")) || 12.5);
  }, []);
  const setEdFsP = v => { setEdFs(v); localStorage.setItem("lake_editor_fs", v); };
  React.useEffect(() => {
    if (!full) return;
    const onKey = e => { if (e.key === "Escape") setFull(false); };
    window.addEventListener("keydown", onKey);
    document.body.style.overflow = "hidden";
    return () => { window.removeEventListener("keydown", onKey); document.body.style.overflow = ""; };
  }, [full]);

  // 卸載時把尚未寫回的變更 flush 掉
  React.useEffect(() => () => {
    if (timer.current) clearTimeout(timer.current);
    if (pending.current) updateArticle(id, pending.current).catch(() => {});
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (!draft) return <div className="a-dim mono text-[12px]">找不到文章。<button className="a-btn ml-2" onClick={back}>返回</button></div>;
  const a = draft;

  // debounce 寫回：累積的 patch 合併送出
  const scheduleSave = patch => {
    pending.current = { ...(pending.current || {}), ...patch };
    setSaving(true);
    if (timer.current) clearTimeout(timer.current);
    timer.current = setTimeout(async () => {
      const toSend = pending.current;
      pending.current = null;
      try { await updateArticle(id, toSend); } catch {}
      setSaving(false);
    }, SAVE_DELAY);
  };

  // 更新本機 draft（即時）＋ 排程寫回（debounce）
  const set = (k, v) => { setDraft(d => ({ ...d, [k]: v })); scheduleSave({ [k]: v }); };

  // 狀態切換：立即寫回（非輸入連打）
  const toggleStatus = () => {
    const next = a.status === "published" ? "draft" : "published";
    setDraft(d => ({ ...d, status: next }));
    if (timer.current) clearTimeout(timer.current);
    const patch = { ...(pending.current || {}), status: next };
    pending.current = null;
    updateArticle(id, patch).catch(() => {});
  };

  const cats = Array.from(new Set([...((db.settings && db.settings.catsBlog) || []), a.cat].filter(Boolean)));
  const wCats = (db.settings && db.settings.catsWorks) || [];
  const gallery = a.gallery || [];

  // gallery 變更：更新本機 draft 後排程寫回整個 gallery 陣列
  const commitGallery = next => { setDraft(d => ({ ...d, gallery: next })); scheduleSave({ gallery: next }); };
  const setGal = (i, v) => { const g = [...gallery]; g[i] = v; commitGallery(g); };
  const addGal = () => { if (gallery.length < 10) commitGallery([...gallery, ""]); };
  const delGal = i => { const g = [...gallery]; g.splice(i, 1); commitGallery(g); };
  const moveGal = (i, dir) => {
    const j = i + dir;
    if (j < 0 || j >= gallery.length) return;
    const g = [...gallery];
    [g[i], g[j]] = [g[j], g[i]];
    commitGallery(g);
  };

  return (
    <div>
      <div className="flex items-center gap-2 mb-4 flex-wrap">
        <button className="a-btn ntc text-[11px]" onClick={back}>← 文章列表</button>
        <div className="ntc font-black text-[15px] flex-1 truncate">編輯：{a.title}</div>
        <StatusPill st={a.status} />
        <button className="a-btn a-btn-acc ntc text-[11px]" onClick={toggleStatus}>
          {a.status === "published" ? "轉為草稿" : "發布上線"}
        </button>
      </div>

      <div className="grid grid-cols-12 gap-4 mb-4">
        <div className="col-span-12 sm:col-span-6">
          <label className="block ntc text-[11px] font-bold mb-1">標題</label>
          <input className="a-inp ntc" value={a.title} onChange={e => set("title", e.target.value)} />
        </div>
        <div className="col-span-6 sm:col-span-3">
          <label className="block ntc text-[11px] font-bold mb-1">分類</label>
          <select className="a-inp ntc" value={a.cat} onChange={e => set("cat", e.target.value)}>
            {cats.map(c => <option key={c}>{c}</option>)}
          </select>
        </div>
        <div className="col-span-6 sm:col-span-3">
          <label className="block ntc text-[11px] font-bold mb-1">日期</label>
          <input className="a-inp mono text-[12px]" value={a.date} onChange={e => set("date", e.target.value)} />
        </div>
        <div className="col-span-12">
          <label className="block ntc text-[11px] font-bold mb-1">附註（顯示於標題下方，如得獎紀錄、技術棧；留空則不顯示）</label>
          <input className="a-inp ntc" value={a.award || ""} placeholder="例：第 39 屆淡江大學五虎崗文學獎，小說組" onChange={e => set("award", e.target.value)} />
        </div>
        <div className="col-span-12">
          <label className="block ntc text-[11px] font-bold mb-1">摘要（列表顯示）</label>
          <input className="a-inp ntc" value={a.blurb} onChange={e => set("blurb", e.target.value)} />
        </div>
        <div className="col-span-12 sm:col-span-8">
          <label className="block ntc text-[11px] font-bold mb-1">橫幅圖片 URL（Notion 式封面，留空則無）</label>
          <input className="a-inp mono text-[12px]" value={a.banner || ""} placeholder="/assets/… 或 https://…" onChange={e => set("banner", e.target.value)} />
        </div>
        <div className="col-span-12 sm:col-span-4 flex items-end pb-1">
          <label className="flex items-center gap-2 ntc text-[12px] cursor-pointer tap">
            <input type="checkbox" checked={!!a.bannerFade} onChange={e => set("bannerFade", e.target.checked)} />
            橫幅半透明淡出
          </label>
        </div>
        {a.banner && (
          <div className="col-span-12">
            <div className="relative overflow-hidden h-24 border a-line" style={{ borderRadius: "var(--a-rad)" }}>
              <img src={a.banner} alt="" className="w-full h-full object-cover" style={a.bannerFade ? { opacity: .45 } : {}}
                onError={e => { e.currentTarget.style.display = "none"; }} />
              <div className="absolute bottom-1 right-2 a-dim mono text-[9px]">banner 預覽</div>
            </div>
          </div>
        )}

        {/* 作品展示連動 */}
        <div className="col-span-12">
          <div className="border a-line p-4 space-y-3" style={{ borderRadius: "var(--a-rad)" }}>
            <div className="flex items-center gap-3 flex-wrap">
              <label className="flex items-center gap-2 ntc text-[12.5px] font-bold cursor-pointer tap">
                <input type="checkbox" checked={!!a.showcase} onChange={e => set("showcase", e.target.checked)} />
                顯示於「作品展示」
              </label>
              {a.showcase && (
                <select className="a-inp ntc" style={{ width: "auto" }} value={a.workCat || wCats[0] || ""}
                  onChange={e => set("workCat", e.target.value)}>
                  {wCats.map(c => <option key={c}>{c}</option>)}
                </select>
              )}
              <div className="a-dim mono text-[10px]">需為已發布才會出現在前台展示。</div>
            </div>
            {a.showcase && (
              <div>
                <div className="flex items-center gap-2 mb-2">
                  <div className="ntc text-[11px] font-bold">展示用封面圖（Gallery 左右呈現，最多 10 張，不顯示於內文）</div>
                  <div className="a-dim mono text-[10px]">{gallery.length} / 10</div>
                  <div className="flex-1"></div>
                  <button className="a-btn ntc text-[10.5px]" onClick={addGal} disabled={gallery.length >= 10}
                    style={gallery.length >= 10 ? { opacity: .4, cursor: "not-allowed" } : {}}>＋新增封面</button>
                </div>
                <div className="space-y-1.5">
                  {gallery.map((g, i) => (
                    <div key={i} className="flex items-center gap-1.5">
                      <span className="a-dim mono text-[10px] w-5 text-right">{i + 1}</span>
                      <div className="w-12 h-9 shrink-0 overflow-hidden border a-line relative" style={{ borderRadius: "var(--a-rad)" }}>
                        {g ? <img src={g} alt="" className="w-full h-full object-cover" onError={e => { e.currentTarget.style.display = "none"; }} />
                           : <div className="a-dim mono text-[8px] flex items-center justify-center h-full">無圖</div>}
                      </div>
                      <input className="a-inp mono text-[11.5px]" value={g} placeholder="/assets/… 或 https://…" onChange={e => setGal(i, e.target.value)} />
                      <button className="a-dim mono text-[10px] cursor-pointer hover:opacity-70 px-1" onClick={() => moveGal(i, -1)}>▲</button>
                      <button className="a-dim mono text-[10px] cursor-pointer hover:opacity-70 px-1" onClick={() => moveGal(i, 1)}>▼</button>
                      <button className="a-btn ntc text-[10.5px]" onClick={() => delGal(i)}>刪除</button>
                    </div>
                  ))}
                  {gallery.length === 0 && <div className="a-dim mono text-[10.5px]">尚未設定封面 — 前台會顯示佔位圖。</div>}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      <div className={full ? "fixed inset-0 z-50 p-4 sm:p-6 flex flex-col" : ""} style={full ? { background: "var(--a-bg)" } : {}}>
        <div className="flex flex-wrap items-center gap-2 mb-2">
          <div className="ntc text-[11px] font-bold">內文 · Markdown{full && <span className="a-dim mono text-[9.5px] ml-2">— {a.title}</span>}</div>
          {!full && <div className="a-dim mono text-[9.5px] hidden md:block">## 標題 · **粗體** · - 列表 · &gt; 引言 · ![圖說](網址)</div>}
          <div className="flex-1"></div>
          <button className="a-btn ntc text-[10.5px]" onClick={() => set("body", (a.body || "") + "\n\n![圖片說明](圖片網址)\n")}>＋插入圖片</button>
          <select className="a-inp mono text-[10.5px] cursor-pointer" title="編輯字級" style={{ width: "auto", padding: "4px 8px", minHeight: "0" }}
            value={edFs} onChange={e => setEdFsP(Number(e.target.value))}>
            <option value={12.5}>字級 · 小</option>
            <option value={14}>字級 · 中</option>
            <option value={16}>字級 · 大</option>
            <option value={18}>字級 · 特大</option>
          </select>
          <button className={"mono text-[10px] px-2 py-1 cursor-pointer tap " + (showPrev ? "a-acc-bg" : "a-dim border a-line")}
            style={{ borderRadius: "var(--a-rad)" }} onClick={() => setShowPrev(p => !p)}>預覽 {showPrev ? "ON" : "OFF"}</button>
          <button className={"mono text-[10px] px-2 py-1 cursor-pointer tap " + (full ? "a-acc-bg" : "a-dim border a-line")}
            style={{ borderRadius: "var(--a-rad)" }} onClick={() => setFull(f => !f)} title={full ? "離開全螢幕（Esc）" : "全螢幕編輯"}>
            {full ? "✕ 離開全螢幕" : "⛶ 全螢幕編輯"}
          </button>
        </div>
        <div className={"grid gap-4 " + (showPrev ? "grid-cols-1 lg:grid-cols-2" : "grid-cols-1") + (full ? " flex-1 min-h-0" : "")}
          style={full ? { gridTemplateRows: showPrev ? undefined : "minmax(0,1fr)", gridAutoRows: "minmax(0,1fr)" } : {}}>
          <textarea className={"a-inp mono leading-[1.8]" + (full ? " h-full min-h-0 resize-none" : "")} rows={full ? undefined : 18} value={a.body}
            style={{ fontSize: edFs + "px" }}
            onChange={e => set("body", e.target.value)} placeholder="開始寫作…"></textarea>
          {showPrev && (
            <div className="a-panel p-5 overflow-y-auto" style={full ? { height: "100%" } : { maxHeight: "480px" }}>
              <div className="a-dim mono text-[9.5px] uppercase tracking-[.2em] mb-3">— Preview —</div>
              <h1 className="ntc text-[22px] font-black tracking-tight mb-2">{a.title}</h1>
              <Markdown text={a.body || "*（空白內文）*"} size={edFs > 14 ? edFs : undefined} />
            </div>
          )}
        </div>
      </div>
      <div className="a-dim mono text-[10px] mt-3">{saving ? "儲存中…" : "✓ 已同步 — 變更經後端 API 儲存，前台即時反映。"}</div>
    </div>
  );
}
