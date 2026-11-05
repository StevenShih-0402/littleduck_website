// frontend/components/front/works-page.jsx
// 作品展示頁：與部落格連動的展示卡片（MAG 網格 / TERM 清單），點開即閱讀全文

"use client";

import React from "react";
import { useRouter } from "next/navigation";
import { useLake } from "@/store/lake-context";
import { SecHead, Chip, Gallery } from "@/components/ui";

function WorkGridCard({ a, idx }) {
  const router = useRouter();
  return (
    <div className="v-card overflow-hidden flex flex-col theme-fade cursor-pointer group" onClick={() => router.push("/blog/" + a.id)}>
      <Gallery images={a.gallery} label={"cover · " + a.title} className="h-40" />
      <div className="p-4 flex-1 flex flex-col">
        <div className="flex items-baseline justify-between gap-2 mb-1">
          <div className="ntc font-black text-[15px] leading-snug group-hover:underline">{a.title}</div>
          <div className="mono text-[10px] v-dim shrink-0">{a.y || ""}</div>
        </div>
        {a.award && <div className="mono text-[9.5px] v-acc mb-2">{a.award}</div>}
        <div className="ntc text-[12px] v-dim leading-relaxed flex-1 line-clamp-3">{a.blurb}</div>
        <div className="mono text-[9.5px] v-dim mt-3 pt-2 border-t v-soft flex justify-between">
          <span>NO. {String(idx + 1).padStart(2, "0")}</span>
          <span className="v-acc">閱讀全文 →</span>
        </div>
      </div>
    </div>
  );
}

function WorkListRow({ a }) {
  const router = useRouter();
  return (
    <div className="v-card p-3 flex gap-3 items-stretch theme-fade cursor-pointer group" onClick={() => router.push("/blog/" + a.id)}>
      <Gallery images={a.gallery} label="cover" className="w-24 sm:w-32 shrink-0 self-stretch min-h-[72px]" />
      <div className="flex-1 min-w-0 py-0.5">
        <div className="flex items-baseline gap-2 flex-wrap">
          <span className="mono text-[10px] v-dim">{a.y}</span>
          <span className="ntc font-bold text-[14px] group-hover:underline">{a.title}</span>
          <span className="mono text-[9.5px] v-acc">{a.award}</span>
        </div>
        <div className="ntc text-[12px] v-dim leading-relaxed mt-1 line-clamp-2">{a.blurb}</div>
        <div className="mono text-[9.5px] v-acc mt-1.5">閱讀全文 →</div>
      </div>
    </div>
  );
}

export default function WorksPage() {
  const { db, theme } = useLake();
  const wCats = db.settings.catsWorks || [];
  const [tab, setTab] = React.useState("全部");
  const all = db.articles.filter(a => a.showcase && a.status === "published");
  const grid = theme !== "night"; // MAG=grid, TERM=list
  const showCats = tab === "全部" ? wCats : [tab];

  return (
    <div>
      <SecHead num="02" zh="作品展示" en="WORKS"
        sub={grid ? "編輯台上的陳列架 — 點開任一作品即可閱讀全文。" : "$ ls works/ --link=blog — 點開作品即閱讀全文。"} />
      <div className="flex flex-wrap gap-2 mb-7">
        <Chip active={tab === "全部"} onClick={() => setTab("全部")}>全部</Chip>
        {wCats.map(c => <Chip key={c} active={tab === c} onClick={() => setTab(c)}>{c}</Chip>)}
      </div>

      {showCats.map(c => {
        const list = all.filter(a => a.workCat === c);
        if (list.length === 0) return null;
        return (
          <section key={c} className="mb-10">
            <div className="mono text-[10.5px] v-dim uppercase tracking-[.2em] mb-3 border-b v-soft pb-1.5">
              {grid ? "— " + c + " —" : "## " + c + "/"} <span className="v-acc">{list.length}</span>
            </div>
            {grid ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {list.map((a, i) => <WorkGridCard key={a.id} a={a} idx={i} />)}
              </div>
            ) : (
              <div className="space-y-2.5">
                {list.map(a => <WorkListRow key={a.id} a={a} />)}
              </div>
            )}
          </section>
        );
      })}
      {all.length === 0 && <div className="mono text-[12px] v-dim py-10 text-center">尚無展示中的作品 — 到後台將文章加入作品展示。</div>}
    </div>
  );
}
