// frontend/store/lake.js
// 業務資料已改接後端 API（見 lib/api.js 與 store/lake-context.jsx），此檔不再是資料來源。
// 僅保留少數 UI 用的輕量 helper（本機暫時 id、顯示用日期）。
// 注意：本檔刻意與設計專案 project/v2/store.js 分歧（SDP-103），見 design-sync/manifest.json。

export const Lake = {
  // 本機暫用 id（僅供尚未取得伺服器 id 前的過場；正式 id 由後端回傳）
  uid() { return Math.random().toString(36).slice(2, 9); },
  // 顯示用日期（前台/後台送出時的佔位；伺服器會以自身時間為準）
  today() {
    const d = new Date();
    const p = n => String(n).padStart(2, "0");
    return d.getFullYear() + " / " + p(d.getMonth() + 1) + " / " + p(d.getDate());
  },
};
