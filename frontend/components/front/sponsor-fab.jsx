// frontend/components/front/sponsor-fab.jsx
// 贊助浮動按鈕與贊助視窗（蜂蜜檸檬）：金額選擇、自訂金額、留言與感謝畫面

"use client";

import React from "react";
import { useLake } from "@/store/lake-context";
import { Lake } from "@/store/lake";

export function HoneyLemonIcon({ size }) {
  return (
    <svg width={size || 26} height={size || 26} viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <path d="M13.5 2 L10.8 8" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
      <path d="M6.5 8 H17.5 L16.3 20.2 A1.8 1.8 0 0 1 14.5 21.8 H9.5 A1.8 1.8 0 0 1 7.7 20.2 Z" stroke="currentColor" strokeWidth="1.6" strokeLinejoin="round" />
      <path d="M7.3 12.5 H16.7 L16 19.6 A1 1 0 0 1 15 20.5 H9 A1 1 0 0 1 8 19.6 Z" fill="currentColor" opacity=".45" />
      <circle cx="17.2" cy="7.2" r="3" stroke="currentColor" strokeWidth="1.4" fill="none" />
      <path d="M17.2 4.6 V9.8 M14.7 7.2 H19.7 M15.5 5.4 L18.9 9 M18.9 5.4 L15.5 9" stroke="currentColor" strokeWidth=".8" />
    </svg>
  );
}

export default function SponsorFab() {
  const { db, update, theme } = useLake();
  const sp = db.settings.sponsor || { title: "請小鴨喝杯蜂蜜檸檬", subtitle: "", tiers: [] };
  const tiers = sp.tiers || [];
  const [open, setOpen] = React.useState(false);
  const [amount, setAmount] = React.useState(tiers[0] ? tiers[0].amt : 50);
  const [custom, setCustom] = React.useState("");
  const [msg, setMsg] = React.useState("");
  const [thanks, setThanks] = React.useState(false);
  const finalAmt = custom !== "" ? Number(custom) || 0 : amount;

  const confirm = () => {
    if (finalAmt <= 0) return;
    update(d => {
      d.donations = d.donations || [];
      d.donations.unshift({ id: Lake.uid(), date: Lake.today(), amt: finalAmt, name: "匿名訪客", note: msg.trim() });
      return d;
    });
    setThanks(true);
    setTimeout(() => { setThanks(false); setOpen(false); setCustom(""); setMsg(""); setAmount(tiers[0] ? tiers[0].amt : 50); }, 2200);
  };

  return (
    <React.Fragment>
      <button onClick={() => setOpen(true)} aria-label={sp.title}
        className="fixed bottom-5 right-5 z-50 w-14 h-14 rounded-full v-acc-bg flex items-center justify-center cursor-pointer hover:scale-105 transition-transform"
        style={{ boxShadow: "0 4px 18px color-mix(in srgb, var(--acc) 45%, transparent)" }}>
        <HoneyLemonIcon />
      </button>

      {open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center px-4" onClick={() => !thanks && setOpen(false)}
          style={{ background: "color-mix(in srgb, var(--ink) 40%, transparent)" }}>
          <div className="v-paper border v-line rad w-full max-w-[380px] p-6 theme-fade" onClick={e => e.stopPropagation()}
            style={{ borderWidth: "var(--heavy)" }}>
            {thanks ? (
              <div className="text-center py-8">
                <div className="v-acc mb-3 flex justify-center"><HoneyLemonIcon size={44} /></div>
                <div className="ntc font-black text-[18px]">感謝你的蜂蜜檸檬！</div>
                <div className="mono text-[11px] v-dim mt-1">NT$ {finalAmt} · {theme === "night" ? "payment ok (200)" : "心意已送達"}</div>
              </div>
            ) : (
              <React.Fragment>
                <div className="flex items-center gap-3 mb-1">
                  <span className="v-acc"><HoneyLemonIcon size={30} /></span>
                  <div className="ntc font-black text-[17px]">{sp.title}</div>
                </div>
                {sp.subtitle && <div className="ntc text-[12px] v-dim mb-1.5">{sp.subtitle}</div>}
                <div className="mono text-[10.5px] v-dim mb-5">{theme === "night" ? "$ donate --to=duck" : "SUPPORT · 贊助創作"}</div>
                <div className="grid grid-cols-3 gap-2 mb-3">
                  {tiers.map(t => (
                    <button key={t.amt} onClick={() => { setAmount(t.amt); setCustom(""); }}
                      className={"tap py-2.5 border rad cursor-pointer flex flex-col items-center gap-1 " +
                        (custom === "" && amount === t.amt ? "v-acc-bg border-transparent" : "v-line hover:opacity-75")}>
                      <span className="ntc font-black text-[15px] leading-none">${t.amt}</span>
                      <span className={"ntc text-[10px] leading-none " + (custom === "" && amount === t.amt ? "opacity-80" : "v-dim")}>{t.label}</span>
                    </button>
                  ))}
                </div>
                <input type="number" min="1" className="inp mono mb-3" placeholder="自訂金額（NT$）"
                  value={custom} onChange={e => setCustom(e.target.value)} />
                <textarea className="inp ntc mb-5" rows={2} maxLength={100} placeholder="想對小鴨說的話（可留空）"
                  value={msg} onChange={e => setMsg(e.target.value)}></textarea>
                <div className="flex gap-2">
                  <button className="btn ntc flex-1" onClick={() => setOpen(false)}>下次一定</button>
                  <button className="btn btn-acc ntc flex-1" onClick={confirm}>贊助 NT$ {finalAmt || "—"}</button>
                </div>
              </React.Fragment>
            )}
          </div>
        </div>
      )}
    </React.Fragment>
  );
}
