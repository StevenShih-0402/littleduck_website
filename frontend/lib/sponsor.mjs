// frontend/lib/sponsor.mjs
// 贊助純函式：金額上下限驗證、結果頁顯示狀態、輪詢間隔。

/** 與後端 PaymentService 的金額上下限同值（新台幣元） */
export const SPONSOR_AMOUNT_MIN = 10;
export const SPONSOR_AMOUNT_MAX = 20000;

/**
 * 驗證贊助金額：須為 SPONSOR_AMOUNT_MIN～SPONSOR_AMOUNT_MAX（含）的整數。
 * @param {unknown} amount
 * @returns {{ ok: boolean, message?: string }}
 */
export function validateSponsorAmount(amount) {
  const range = "金額需為 " + SPONSOR_AMOUNT_MIN + "～" + SPONSOR_AMOUNT_MAX + " 元的整數";
  if (typeof amount !== "number" || !Number.isInteger(amount)) return { ok: false, message: range };
  if (amount < SPONSOR_AMOUNT_MIN || amount > SPONSOR_AMOUNT_MAX) return { ok: false, message: range };
  return { ok: true };
}

/**
 * 付款狀態對應的結果頁顯示。
 * @param {unknown} status
 * @returns {"paid" | "pending" | "failed" | "missing"}
 */
export function resultView(status) {
  if (status === "paid" || status === "pending" || status === "failed") return status;
  return "missing";
}

/** 輪詢延遲（毫秒），依序遞增；次數用盡回 null */
const POLL_DELAYS = [2000, 3000, 5000, 5000, 8000, 8000, 10000, 10000];

/**
 * @param {number} attempt 從 0 起算的已輪詢次數
 * @returns {number | null}
 */
export function nextPollDelay(attempt) {
  if (!Number.isInteger(attempt) || attempt < 0 || attempt >= POLL_DELAYS.length) return null;
  return POLL_DELAYS[attempt];
}
