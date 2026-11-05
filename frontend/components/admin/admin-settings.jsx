// frontend/components/admin/admin-settings.jsx
// 網站設定：基本資料、部落格/作品分類管理、社群連結與示範資料重設

"use client";

import React from "react";
import { useLake } from "@/store/lake-context";
import { Lake } from "@/store/lake";

export default function AdminSettings() {
  const { db, update } = useLake();
  const s = db.settings;
  const set = (k, v) => update(d => { d.settings[k] = v; return d; });
  const setSocial = (i, k, v) => update(d => { d.settings.social[i][k] = v; return d; });
  const addSocial = () => update(d => { d.settings.social.push({ label: "新連結", url: "" }); return d; });
  const delSocial = i => update(d => { d.settings.social.splice(i, 1); return d; });
  const [saved, setSaved] = React.useState(false);
  const [newCatB, setNewCatB] = React.useState("");
  const [newCatW, setNewCatW] = React.useState("");
  const flash = () => { setSaved(true); setTimeout(() => setSaved(false), 2000); };

  const addCat = (key, val, clear) => {
    const v = val.trim();
    if (!v) return;
    update(d => {
      if (!d.settings[key].includes(v)) d.settings[key].push(v);
      return d;
    });
    clear("");
  };
  const delCat = (key, c) => {
    const inUse = key === "catsBlog"
      ? db.articles.some(a => a.cat === c)
      : db.articles.some(a => a.showcase && a.workCat === c);
    if (inUse) { alert("尚有文章使用「" + c + "」分類，請先調整那些文章。"); return; }
    update(d => { d.settings[key] = d.settings[key].filter(x => x !== c); return d; });
  };

  const inp = (k, extra) => (
    <input className="a-inp" value={s[k]} onChange={e => set(k, e.target.value)} {...(extra || {})} />
  );

  const CatEditor = (key, label, hint, newVal, setNewVal) => (
    <div>
      <div className="ntc text-[12px] font-bold mb-1.5">{label}</div>
      <div className="flex flex-wrap gap-1.5 mb-2">
        {s[key].map(c => (
          <span key={c} className="inline-flex items-center gap-1.5 border a-line px-2.5 py-1 ntc text-[12px]" style={{ borderRadius: "var(--a-rad)" }}>
            {c}
            <button className="a-dim mono text-[11px] cursor-pointer hover:opacity-70" title="刪除分類" onClick={() => delCat(key, c)}>✕</button>
          </span>
        ))}
      </div>
      <div className="flex gap-2">
        <input className="a-inp ntc" style={{ maxWidth: "200px" }} value={newVal} placeholder="新分類名稱"
          onChange={e => setNewVal(e.target.value)}
          onKeyDown={e => e.key === "Enter" && addCat(key, newVal, setNewVal)} />
        <button className="a-btn ntc text-[11px]" onClick={() => addCat(key, newVal, setNewVal)}>＋新增</button>
      </div>
      <div className="a-dim text-[10px] mono mt-1">{hint}</div>
    </div>
  );

  return (
    <div className="max-w-[680px] space-y-5">
      <div className="a-panel p-5 space-y-4">
        <div className="ntc font-black text-[14px]">基本資料</div>
        <div><label className="block ntc text-[12px] font-bold mb-1">網站標題</label>{inp("title")}</div>
        <div><label className="block ntc text-[12px] font-bold mb-1">英文副標</label>{inp("titleEn")}</div>
        <div>
          <label className="block ntc text-[12px] font-bold mb-1">a.k.a. 稱號（首頁名片顯示）</label>
          {inp("aka", { placeholder: "例：站長小鴨" })}
        </div>
        <div><label className="block ntc text-[12px] font-bold mb-1">標語（首頁大標下方，等寬字）</label>{inp("tagline", { placeholder: "例：Programmer × Novelist" })}</div>
        <div><label className="block ntc text-[12px] font-bold mb-1">副標題（接在標語後方）</label>{inp("subtitle", { placeholder: "例：程式與小說，是成本最低的『創作』。" })}</div>
        <div><label className="block ntc text-[12px] font-bold mb-1">座右銘</label>{inp("motto")}</div>
        <div>
          <label className="block ntc text-[12px] font-bold mb-1">公告列</label>
          {inp("announcement")}
          <div className="a-dim text-[10px] mono mt-1">留空即隱藏前台公告列。改動立即反映在前台。</div>
        </div>
      </div>

      <div className="a-panel p-5 space-y-4">
        <div className="ntc font-black text-[14px]">分類管理</div>
        {CatEditor("catsBlog", "部落格分類", "新分類會出現在文章編輯器的分類選單與前台篩選。使用中的分類無法刪除。", newCatB, setNewCatB)}
        {CatEditor("catsWorks", "作品展示分類", "作品展示頁依此分組；文章編輯器勾選展示後可選擇分類。", newCatW, setNewCatW)}
      </div>

      <div className="a-panel p-5 space-y-3">
        <div className="flex items-center">
          <div className="ntc font-black text-[14px] flex-1">社群連結</div>
          <button className="a-btn ntc text-[11px]" onClick={addSocial}>＋新增連結</button>
        </div>
        {s.social.map((so, i) => (
          <div key={i} className="grid grid-cols-[140px_1fr_auto] gap-2">
            <input className="a-inp" value={so.label} onChange={e => setSocial(i, "label", e.target.value)} />
            <input className="a-inp mono text-[12px]" value={so.url} placeholder="https://…" onChange={e => setSocial(i, "url", e.target.value)} />
            <button className="a-btn ntc text-[11px] a-dim" title="刪除" onClick={() => delSocial(i)}>✕</button>
          </div>
        ))}
        {s.social.length === 0 && <div className="a-dim mono text-[11px]">尚無連結 — 頁尾將不顯示社群區。</div>}
      </div>

      <div className="a-panel p-4 a-dim ntc text-[12px]">
        原型注記 — 目前所有內容（文章、作品、留言、設定、贊助）都存在瀏覽器儲存中模擬資料庫；正式版換成後端 API 後，這些後台介面維持不變，直接對接資料庫即可。
      </div>

      <div className="flex items-center gap-3">
        <button className="a-btn a-btn-acc ntc" onClick={flash}>儲存設定</button>
        {saved && <span className="mono text-[11px] a-acc">✓ 已儲存（原型即時生效）</span>}
        <div className="flex-1"></div>
        <button className="a-btn ntc text-[11px]" onClick={() => { if (confirm("重設所有內容為預設值？")) { Lake.reset(); location.reload(); } }}>重設示範資料</button>
      </div>
    </div>
  );
}
