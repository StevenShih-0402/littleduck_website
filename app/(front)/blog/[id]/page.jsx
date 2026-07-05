// frontend/app/(front)/blog/[id]/page.jsx
// 文章閱讀路由：以動態參數 id 渲染 ArticlePage

"use client";

import ArticlePage from "@/components/front/article-page";

export default function Page({ params }) {
  return <ArticlePage id={params.id} />;
}
