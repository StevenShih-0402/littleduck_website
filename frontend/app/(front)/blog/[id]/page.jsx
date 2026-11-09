// frontend/app/(front)/blog/[id]/page.jsx
// 文章閱讀路由：以動態參數 id 渲染 ArticlePage

"use client";

import { useParams } from "next/navigation";

import ArticlePage from "@/components/front/article-page";

// Next 15 起 page 的 params 改為 Promise，client component 改以 useParams() 同步取得路由參數
export default function Page() {
  const { id } = useParams();
  return <ArticlePage id={id} />;
}
