// frontend/components/admin/status-pill.jsx
// 文章狀態標籤：已發布 / 草稿

"use client";

import React from "react";

export default function StatusPill({ st }) {
  const on = st === "published";
  return (
    <span className="mono text-[9.5px] px-2 py-0.5 border a-line inline-flex items-center gap-1.5" style={{ borderRadius: "var(--a-rad)" }}>
      <span className="w-1.5 h-1.5 rounded-full" style={{ background: on ? "var(--a-acc)" : "var(--a-dim)" }}></span>
      {on ? "已發布" : "草稿"}
    </span>
  );
}
