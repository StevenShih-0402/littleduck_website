// frontend/lib/feature-flags.mjs
// 功能開關純函式：key 常數、嚴格判斷（fail-closed）、前台顯示模式。

/** 開關 key 常數，需與後端登記表一致 */
export const FEATURES = Object.freeze({
  SPONSOR: "sponsor",
  GUESTBOOK: "guestbook",
  ARTICLE_COMMENTS: "article_comments",
});

/**
 * 只有值嚴格為 true 才算啟用；flags 為空、缺 key、非布林值一律視為停用。
 * @param {Record<string, unknown> | null | undefined} flags
 * @param {string} key
 * @returns {boolean}
 */
export function isFeatureEnabled(flags, key) {
  if (!flags || typeof flags !== "object") return false;
  return flags[key] === true;
}

/**
 * 前台區塊顯示模式：啟用 show；停用且為管理員預覽 preview；其他 hide。
 * @param {Record<string, unknown> | null | undefined} flags
 * @param {string} key
 * @param {boolean} adminPreview
 * @returns {"show" | "preview" | "hide"}
 */
export function gateMode(flags, key, adminPreview) {
  if (isFeatureEnabled(flags, key)) return "show";
  return adminPreview ? "preview" : "hide";
}
