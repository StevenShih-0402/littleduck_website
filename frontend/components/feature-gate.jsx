// frontend/components/feature-gate.jsx
// 功能開關閘門：啟用時照常渲染；停用且為管理員預覽時以虛線外框加「停用中・預覽」小標渲染；其他情況隱藏。
// 開關尚未載入完成時回 null，避免閃現。

"use client";

import React from "react";
import { useLake } from "@/store/lake-context";
import { gateMode } from "@/lib/feature-flags.mjs";

export function PreviewBadge() {
  return (
    <span className="mono v-acc text-[10px] px-1.5 py-0.5 border v-line rad v-paper whitespace-nowrap">
      停用中・預覽
    </span>
  );
}

/**
 * @param {{ flag: string, children: React.ReactNode | ((preview: boolean) => React.ReactNode), inline?: boolean }} props
 * inline 為 true 時不包外框（由 children 函式依 preview 自行顯示小標，供固定定位元件使用）
 */
export default function FeatureGate({ flag, children, inline }) {
  const { features, featuresLoaded, adminPreview } = useLake();
  if (!featuresLoaded) return null;
  const mode = gateMode(features, flag, adminPreview);
  if (mode === "hide") return null;
  const preview = mode === "preview";
  if (inline || typeof children === "function") {
    return <React.Fragment>{typeof children === "function" ? children(preview) : children}</React.Fragment>;
  }
  if (!preview) return <React.Fragment>{children}</React.Fragment>;
  return (
    <div className="relative border border-dashed v-line rad p-3 pt-5 my-2">
      <div className="absolute -top-2.5 right-3"><PreviewBadge /></div>
      {children}
    </div>
  );
}
