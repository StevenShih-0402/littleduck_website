// frontend/app/(front)/sponsor/result/page.jsx
// 贊助付款結果頁路由：綠界付款完成後導回，依 trade_no 查詢結果

"use client";

import React from "react";
import SponsorResultPage from "@/components/front/sponsor-result-page";

export default function Page() {
  return (
    <React.Suspense fallback={null}>
      <SponsorResultPage />
    </React.Suspense>
  );
}
