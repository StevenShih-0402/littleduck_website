// 將使用者可輸入的網址分類為站內、外部或不可點（SDP-105，規則見 design.md D2）。
// 只有站內路徑與 http/https/mailto 會成為可點擊連結，其餘一律不產生 href。
const EXTERNAL_SCHEMES = new Set(["http", "https", "mailto"]);

/**
 * @param {unknown} raw
 * @returns {{ kind: "internal" | "external" | "none", href: string }}
 */
export function classifyHref(raw) {
  const href = typeof raw === "string" ? raw.trim() : "";
  if (!href || href.startsWith("#")) return { kind: "none", href: "" };
  if (href.startsWith("/")) {
    // "//host" 與 "/\host" 會被瀏覽器當成協定相對網址
    if (href[1] === "/" || href[1] === "\\") return { kind: "none", href: "" };
    return { kind: "internal", href };
  }
  const m = href.match(/^([a-zA-Z][a-zA-Z0-9+.-]*):/);
  if (m && EXTERNAL_SCHEMES.has(m[1].toLowerCase())) return { kind: "external", href };
  return { kind: "none", href: "" };
}
