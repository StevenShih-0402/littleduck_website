// frontend/store/lake-context.jsx
// LakeCtx Provider：業務資料改由後端 API 取得（db/loading/error），提供非同步 actions；
// 主題偏好仍持久化於 localStorage。

"use client";

import React from "react";
import {
  apiGet, apiPost, apiPut, apiDelete,
  getToken, setToken, clearToken, setUnauthorizedHandler,
} from "@/lib/api";

const LakeCtx = React.createContext(null);

export function useLake() {
  return React.useContext(LakeCtx);
}

// SSR/載入期安全空值，確保元件讀取不崩潰
const EMPTY_DB = { settings: null, articles: [], comments: [], donations: [], stats: { daily: [] } };

export function LakeProvider({ children }) {
  const [db, setDb] = React.useState(EMPTY_DB);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState(null);
  const [authed, setAuthed] = React.useState(false);
  const [theme, setThemeState] = React.useState("day");

  const patchDb = React.useCallback(fn => setDb(prev => fn(prev)), []);

  // ── 主題偏好（localStorage，與業務資料無關） ──
  const setTheme = t => {
    setThemeState(t);
    try { localStorage.setItem("lake_theme_v2", t); } catch {}
  };
  React.useEffect(() => {
    document.documentElement.classList.toggle("night", theme === "night");
  }, [theme]);

  // ── 公開資料：mount 時並行載入 settings / articles / comments ──
  const loadPublic = React.useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [settings, articles, comments] = await Promise.all([
        apiGet("/settings"),
        apiGet("/articles"),
        apiGet("/comments"),
      ]);
      setDb(prev => ({
        ...prev,
        settings: settings || null,
        articles: articles || [],
        comments: comments || [],
      }));
    } catch (e) {
      setError(e.message || "載入失敗");
    } finally {
      setLoading(false);
    }
  }, []);

  // ── 後台資料：authed 時載入 admin 文章/留言/贊助/統計並合併 ──
  const reloadAdmin = React.useCallback(async () => {
    if (!getToken()) return;
    try {
      const [articles, comments, donations, stats] = await Promise.all([
        apiGet("/admin/articles"),
        apiGet("/admin/comments"),
        apiGet("/donations"),
        apiGet("/stats"),
      ]);
      // stats 後端為 [{day, views, visitors}]，對映到 db.stats.daily
      const daily = (stats || []).map(s => ({ day: s.day, views: s.views, visitors: s.visitors }));
      setDb(prev => ({
        ...prev,
        articles: articles || prev.articles,
        comments: comments || prev.comments,
        donations: donations || [],
        stats: { daily },
      }));
    } catch (e) {
      setError(e.message || "後台資料載入失敗");
    }
  }, []);

  // ── 401 → 強制登出並導回登入 ──
  const logout = React.useCallback(() => {
    clearToken();
    setAuthed(false);
    // 登出後退回只含公開資料（清掉草稿/全部留言/贊助/統計）
    loadPublic();
  }, [loadPublic]);

  React.useEffect(() => {
    setUnauthorizedHandler(() => {
      setAuthed(false);
    });
    return () => setUnauthorizedHandler(null);
  }, []);

  // 初始化：讀主題、判斷是否已有 token、載入公開資料（若有 token 再載後台）
  React.useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- 掛載後才讀 localStorage，避免 SSR hydration 不一致
    try { setThemeState(localStorage.getItem("lake_theme_v2") || "day"); } catch {}
    const hasToken = !!getToken();
    setAuthed(hasToken);
    (async () => {
      await loadPublic();
      if (hasToken) await reloadAdmin();
    })();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // ── Auth actions ──
  const login = React.useCallback(async (username, password) => {
    const data = await apiPost("/auth/login", { username, password });
    if (!data || !data.token) throw new Error("登入回應無效");
    setToken(data.token);
    setAuthed(true);
    await reloadAdmin();
    return data;
  }, [reloadAdmin]);

  // ── Article actions ──
  const createArticle = React.useCallback(async (partial) => {
    const created = await apiPost("/articles", partial);
    setDb(prev => ({ ...prev, articles: [created, ...prev.articles] }));
    return created;
  }, []);

  const updateArticle = React.useCallback(async (id, patch) => {
    const current = db.articles.find(a => a.id === id) || {};
    const updated = await apiPut("/articles/" + id, { ...current, ...patch });
    setDb(prev => ({
      ...prev,
      articles: prev.articles.map(a => (a.id === id ? (updated || { ...a, ...patch }) : a)),
    }));
    return updated;
  }, [db.articles]);

  const deleteArticle = React.useCallback(async (id) => {
    await apiDelete("/articles/" + id);
    setDb(prev => ({ ...prev, articles: prev.articles.filter(a => a.id !== id) }));
  }, []);

  // ── Comment actions ──
  const postComment = React.useCallback(async ({ name, text, articleId }) => {
    const created = await apiPost("/comments", { name, text, articleId: articleId || "" });
    // 新留言為 pending，前台不顯示；仍併入 db 供後台/狀態轉換使用
    if (created) setDb(prev => ({ ...prev, comments: [created, ...prev.comments] }));
    return created;
  }, []);

  const setCommentStatus = React.useCallback(async (id, status) => {
    const updated = await apiPut("/comments/" + id + "/status", { status });
    setDb(prev => ({
      ...prev,
      comments: prev.comments.map(c => (c.id === id ? (updated || { ...c, status }) : c)),
    }));
    return updated;
  }, []);

  const replyComment = React.useCallback(async (id, reply) => {
    const updated = await apiPut("/comments/" + id + "/reply", { reply });
    setDb(prev => ({
      ...prev,
      comments: prev.comments.map(c => (c.id === id ? (updated || { ...c, reply }) : c)),
    }));
    return updated;
  }, []);

  // ── Settings action ──
  const updateSettings = React.useCallback(async (next) => {
    const updated = await apiPut("/settings", next);
    setDb(prev => ({ ...prev, settings: updated || next }));
    return updated;
  }, []);

  // ── Stats：瀏覽回報，fire-and-forget（絕不丟錯） ──
  const reportPageview = React.useCallback((path) => {
    apiPost("/stats/pageview", { path }).catch(() => {});
  }, []);

  const value = {
    db, loading, error, authed, theme, setTheme,
    login, logout, reloadAdmin,
    createArticle, updateArticle, deleteArticle,
    postComment, setCommentStatus, replyComment,
    updateSettings, reportPageview,
  };

  return <LakeCtx.Provider value={value}>{children}</LakeCtx.Provider>;
}
