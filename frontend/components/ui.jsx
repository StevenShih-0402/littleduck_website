// frontend/components/ui.jsx
// 共用 UI 元件：迷你 Markdown 渲染器、圖片佔位、封面 Gallery、區塊標題與分類 Chip

"use client";

import React from "react";

// tiny markdown renderer (headings/bold/italic/code/list/quote/link/hr)
function mdInline(s) {
  const out = [];
  let rest = s, k = 0;
  const re = /(\*\*([^*]+)\*\*)|(\*([^*]+)\*)|(`([^`]+)`)|(\[([^\]]+)\]\(([^)]+)\))/;
  while (rest) {
    const m = rest.match(re);
    if (!m) { out.push(rest); break; }
    if (m.index > 0) out.push(rest.slice(0, m.index));
    if (m[1]) out.push(<strong key={k++}>{m[2]}</strong>);
    else if (m[3]) out.push(<em key={k++}>{m[4]}</em>);
    else if (m[5]) out.push(<code key={k++}>{m[6]}</code>);
    else if (m[7]) out.push(<a key={k++} href={m[9]} onClick={e => e.preventDefault()}>{m[8]}</a>);
    rest = rest.slice(m.index + m[0].length);
  }
  return out;
}

export function Markdown({ text, size }) {
  const blocks = [];
  const lines = (text || "").split("\n");
  let i = 0, k = 0;
  while (i < lines.length) {
    const ln = lines[i];
    if (!ln.trim()) { i++; continue; }
    if (ln.startsWith("### ")) { blocks.push(<h3 key={k++}>{mdInline(ln.slice(4))}</h3>); i++; }
    else if (ln.startsWith("## ")) { blocks.push(<h2 key={k++}>{mdInline(ln.slice(3))}</h2>); i++; }
    else if (ln.startsWith("# ")) { blocks.push(<h2 key={k++}>{mdInline(ln.slice(2))}</h2>); i++; }
    else if (/^---+$/.test(ln.trim())) { blocks.push(<hr key={k++} />); i++; }
    else if (ln.startsWith("> ")) {
      const q = [];
      while (i < lines.length && lines[i].startsWith("> ")) { q.push(lines[i].slice(2)); i++; }
      blocks.push(<blockquote key={k++}>{mdInline(q.join(" "))}</blockquote>);
    }
    else if (/^!\[[^\]]*\]\([^)]*\)\s*$/.test(ln)) {
      const m = ln.match(/^!\[([^\]]*)\]\(([^)]*)\)/);
      blocks.push(<MdImage key={k++} alt={m[1]} src={m[2]} />);
      i++;
    }
    else if (/^[-*] /.test(ln)) {
      const items = [];
      while (i < lines.length && /^[-*] /.test(lines[i])) { items.push(lines[i].slice(2)); i++; }
      blocks.push(<ul key={k++}>{items.map((it, j) => <li key={j}>{mdInline(it)}</li>)}</ul>);
    }
    else {
      const p = [];
      while (i < lines.length && lines[i].trim() && !/^(#|>|[-*] |---)/.test(lines[i])) { p.push(lines[i]); i++; }
      blocks.push(<p key={k++}>{mdInline(p.join(" "))}</p>);
    }
  }
  return <div className="md ntc" style={size ? { fontSize: size + "px" } : undefined}>{blocks}</div>;
}

// markdown image (placeholder fallback when src missing/broken)
export function MdImage({ alt, src }) {
  const [err, setErr] = React.useState(false);
  if (!src || err) return <div className="ph h-40 my-4">{alt || "image"}</div>;
  return (
    <figure className="my-4">
      <img src={src} alt={alt} className="w-full rad border v-line" onError={() => setErr(true)} />
      {alt && <figcaption className="mono text-[10px] v-dim mt-1.5 text-center">▸ {alt}</figcaption>}
    </figure>
  );
}

// thumbnail placeholder
export function Thumb({ label, className }) {
  return <div className={"ph " + (className || "")}>{label}</div>;
}

// cover gallery — 左右切換的展示封面（最多 10 張）
export function Gallery({ images, label, className }) {
  const imgs = (images || []).slice(0, 10);
  const [i, setI] = React.useState(0);
  const [err, setErr] = React.useState({});
  const n = imgs.length;
  if (n === 0) return <Thumb label={label} className={className} />;
  const idx = Math.min(i, n - 1);
  const src = imgs[idx];
  const nav = (e, dir) => { e.stopPropagation(); setI((idx + dir + n) % n); };
  return (
    <div className={"relative overflow-hidden group/g " + (className || "")}>
      {src && !err[idx]
        ? <img src={src} alt={label} className="absolute inset-0 w-full h-full object-cover"
            onError={() => setErr(p => ({ ...p, [idx]: true }))} />
        : <div className="ph absolute inset-0">{label + " · " + (idx + 1) + "/" + n}</div>}
      {n > 1 && (
        <React.Fragment>
          <button aria-label="上一張" onClick={e => nav(e, -1)}
            className="absolute left-1.5 top-1/2 -translate-y-1/2 w-8 h-8 rad border v-line mono text-[13px] cursor-pointer flex items-center justify-center opacity-0 group-hover/g:opacity-100 transition-opacity"
            style={{ background: "color-mix(in srgb, var(--paper) 82%, transparent)", color: "var(--ink)" }}>‹</button>
          <button aria-label="下一張" onClick={e => nav(e, 1)}
            className="absolute right-1.5 top-1/2 -translate-y-1/2 w-8 h-8 rad border v-line mono text-[13px] cursor-pointer flex items-center justify-center opacity-0 group-hover/g:opacity-100 transition-opacity"
            style={{ background: "color-mix(in srgb, var(--paper) 82%, transparent)", color: "var(--ink)" }}>›</button>
          <div className="absolute bottom-1.5 left-0 right-0 flex justify-center gap-1.5">
            {imgs.map((_, j) => (
              <button key={j} aria-label={"第 " + (j + 1) + " 張"} onClick={e => { e.stopPropagation(); setI(j); }}
                className="w-2 h-2 rounded-full border cursor-pointer"
                style={{ background: j === idx ? "var(--acc)" : "color-mix(in srgb, var(--paper) 70%, transparent)", borderColor: "var(--line)" }}></button>
            ))}
          </div>
        </React.Fragment>
      )}
    </div>
  );
}

// front section header — heavy rule in day, thin in night
/** @param {{ num: string, zh: string, en: string, sub?: import('react').ReactNode }} props sub 為選填副標 */
export function SecHead({ num, zh, en, sub }) {
  return (
    <div className="mb-5 theme-fade">
      <div className="flex items-baseline justify-between border-t rule-heavy pt-3">
        <div className="ntc text-[24px] sm:text-[30px] font-black leading-none tracking-tight">
          <span className="v-acc">{num}</span> / {zh}
        </div>
        <div className="mono text-[10px] sm:text-[11px] v-dim uppercase tracking-[.2em]">{en}</div>
      </div>
      {sub && <div className="ntc text-[12.5px] v-dim mt-2 max-w-2xl">{sub}</div>}
    </div>
  );
}

// category chip
export function Chip({ children, active, onClick }) {
  return (
    <button onClick={onClick}
      className={"mono text-[11px] px-2.5 py-1 border rad cursor-pointer tap " + (active ? "v-acc-bg border-transparent" : "v-line v-dim hover:opacity-75")}>
      {children}
    </button>
  );
}
