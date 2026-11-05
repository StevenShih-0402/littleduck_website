// frontend/app/(front)/layout.jsx
// 前台路由群組 layout：所有前台頁面共用 FrontShell 外框

"use client";

import FrontShell from "@/components/front/front-shell";

export default function FrontLayout({ children }) {
  return <FrontShell>{children}</FrontShell>;
}
