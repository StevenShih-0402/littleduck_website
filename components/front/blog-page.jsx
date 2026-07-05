// frontend/components/front/blog-page.jsx
// 部落格列表頁：分類篩選 Chip 與依日期排序的文章列表

"use client";

import React from "react";
import { useRouter } from "next/navigation";
import { useLake } from "@/store/lake-context";
import { SecHead, Chip } from "@/components/ui";

export default function BlogPage() {
  const { db, theme } = useLake();
  const router = useRouter();
  const [cat, setCat] = React.useState("全部");
  const pub = db.articles.filter(a => a.status === "published");
  const cats = ["全部", ...Array.from(new Set([...(db.settings.catsBlog || []), ...pub.map(a => a.cat)]))];
  const list = (cat === "全部" ? pub : pub.filter(a => a.cat === cat))
    .slice().sort((a, b) => b.date < a.date ? -1 : 1);

  return (
    <div>
      <SecHead num="03" zh="部落格" en="BLOG"
        sub={theme === "night" ? "$ ls blog/ --sort=date | grep " + (cat === "全部" ? "'*'" : "'" + cat + "'") : "小說、隨筆與技術筆記。"} />
      <div className="flex flex-wrap gap-2 mb-6">
        {cats.map(c => <Chip key={c} active={cat === c} onClick={() => setCat(c)}>{c}</Chip>)}
      </div>
      <div className="space-y-0">
        {list.map(a => (
          <button key={a.id} onClick={() => router.push("/blog/" + a.id)}
            className="w-full text-left py-4 border-b v-soft cursor-pointer group block">
            <div className="flex items-baseline gap-3 mb-1">
              <span className="mono text-[10px] v-dim shrink-0 w-24">{a.date}</span>
              <span className="mono text-[9.5px] v-acc shrink-0 border v-line rad px-1.5 py-0.5">{a.cat}</span>
              <span className="ntc text-[16px] sm:text-[17px] font-black group-hover:underline">{a.title}</span>
            </div>
            <div className="ntc text-[12.5px] v-dim leading-relaxed sm:pl-[108px] max-w-2xl">{a.blurb}</div>
            {a.award && <div className="mono text-[10px] v-acc2 mt-1 sm:pl-[108px]">★ {a.award}</div>}
          </button>
        ))}
      </div>
      {list.length === 0 && <div className="mono text-[12px] v-dim py-10 text-center">此分類尚無文章。</div>}
    </div>
  );
}
