// frontend/store/lake-context.jsx
// LakeCtx Provider：提供 db 狀態、update、theme 與 setTheme 給全站元件

"use client";

import React from "react";
import { Lake } from "./lake";

const LakeCtx = React.createContext(null);

export function useLake() {
  return React.useContext(LakeCtx);
}

export function LakeProvider({ children }) {
  // SSR 先渲染確定性的 seed，掛載後再讀取 localStorage，避免 hydration 不一致
  const [db, setDb] = React.useState(() => JSON.parse(JSON.stringify(Lake.seed)));
  const [theme, setThemeState] = React.useState("day");

  React.useEffect(() => {
    setDb(Lake.load());
    setThemeState(localStorage.getItem("lake_theme_v2") || "day");
  }, []);

  const update = fn => setDb(prev => {
    const next = fn(JSON.parse(JSON.stringify(prev)));
    Lake.save(next);
    return next;
  });
  const setTheme = t => {
    setThemeState(t);
    localStorage.setItem("lake_theme_v2", t);
  };

  React.useEffect(() => {
    document.documentElement.classList.toggle("night", theme === "night");
  }, [theme]);

  return (
    <LakeCtx.Provider value={{ db, update, theme, setTheme }}>
      {children}
    </LakeCtx.Provider>
  );
}
