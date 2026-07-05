// frontend/store/lake.js
// localStorage 資料層（Lake）：種子資料建立、載入、儲存與重設；正式版換成後端 API 時介面不變

import { SITE } from "@/data/site";

const KEY = "lakeside_v2_store";
const S = SITE;

// deterministic pseudo-random (stable fake stats)
function makeRnd(seed) {
  let x = seed;
  return () => ((x = (x * 1103515245 + 12345) % 2147483648) / 2147483648);
}

function seedStats() {
  const rnd = makeRnd(42);
  const days = [];
  const today = new Date(2026, 6, 2); // 2026/07/02
  for (let i = 29; i >= 0; i--) {
    const d = new Date(today);
    d.setDate(d.getDate() - i);
    const wk = d.getDay() === 0 || d.getDay() === 6 ? 1.4 : 1;
    const views = Math.round((70 + rnd() * 130) * wk);
    const visitors = Math.round(views * (0.42 + rnd() * 0.18));
    days.push({ d: (d.getMonth() + 1) + "/" + d.getDate(), views, visitors });
  }
  return days;
}

const awardMap = {};
(S.literature.awards || []).forEach(a => { awardMap[a.t] = a.sub; });

const rndV = makeRnd(7);

// 小說 → 文章（未私藏者同時上架到作品展示）
const articles = S.literature.articles.map((a, i) => ({
  id: "a" + (i + 1),
  title: a.t,
  cat: "小說",
  date: a.y + " / " + String((i * 3) % 12 + 1).padStart(2, "0") + " / " + String((i * 7) % 26 + 2).padStart(2, "0"),
  y: a.y,
  status: a.priv ? "draft" : "published",
  views: 120 + Math.round(rndV() * 1400),
  blurb: a.blurb,
  award: awardMap[a.t] || "",
  banner: "", bannerFade: false,
  showcase: !a.priv,           // 顯示於作品展示
  workCat: "小說作品",          // 作品展示分類
  gallery: [],                 // 展示用封面圖（最多 10 張，不出現在內文）
  body: "> " + a.blurb + "\n\n（全文整理中，之後由後台補上。）",
}));

// 程式專案 → 文章（與部落格連動：點開作品即閱讀文章）
const projArticles = S.programming.projects.map((p, i) => ({
  id: "p" + (i + 1),
  title: p.t,
  cat: "技術",
  date: (2024 - Math.floor(i / 4)) + " / " + String((i * 2) % 12 + 1).padStart(2, "0") + " / " + String((i * 5) % 26 + 3).padStart(2, "0"),
  y: 2024 - Math.floor(i / 4),
  status: "published",
  views: 60 + Math.round(rndV() * 700),
  blurb: p.blurb,
  award: p.stack || "",
  banner: "", bannerFade: false,
  showcase: true,
  workCat: "程式專案",
  gallery: [],
  body: "> " + p.blurb + "\n\n**技術棧**：" + (p.stack || "—") + (p.link ? "\n\n**連結**：" + p.link : "") + "\n\n（專案介紹整理中，之後由後台補上。）",
}));
articles.push(...projArticles);

// one essay so 分類 isn't only 小說
articles.unshift({
  id: "a0",
  title: "湖畔改版記事：從 Google Sites 到自己的小屋",
  cat: "隨筆",
  date: "2026 / 06 / 21",
  y: 2026,
  status: "published",
  views: 213,
  blurb: "記錄這次網站 2.0 改版的想法：為什麼想要一個結合部落格與作品展示的個人小屋。",
  award: "",
  banner: "/assets/duck-mini.jpg", bannerFade: true,
  showcase: false, workCat: "", gallery: [],
  body: "## 為什麼改版\n\n舊網站是 Google Sites 拼起來的，展示可以，但**寫東西**很痛苦。\n\n這次想要的是：\n\n- 一個能隨手發文的部落格\n- 一個能好好陳列專案與小說的展示區\n- 一個自己的後台，管文章、看數據\n\n![舊站截圖（待補）]()\n\n> 鴨子在湖面上看似悠哉漂泊，實則在底下奮力划水。\n\n之後的隨筆、技術筆記都會發在這裡。",
});

// 給改版記事 & 第一篇小說一點示範封面圖
const demo = articles.find(a => a.id === "a1");
if (demo) demo.gallery = ["/assets/duck-mini.jpg", "", ""];

const seed = {
  version: 4,
  settings: {
    title: S.brand,
    titleEn: S.brandEn,
    owner: S.owner,
    aka: "站長小鴨",
    tagline: "Programmer × Novelist",
    subtitle: "程式與小說，是成本最低的『創作』。",
    motto: S.motto,
    announcement: "小屋 2.0 開張 — 留言板已開放，歡迎簽到。",
    social: [
      { label: "Instagram", url: "#" },
      { label: "GitHub", url: "#" },
      { label: "巴哈姆特", url: "#" },
      { label: "Medium", url: "#" },
    ],
    catsBlog: ["小說", "隨筆", "技術"],
    catsWorks: ["程式專案", "小說作品"],
    sponsor: {
      title: "請小鴨喝杯蜂蜜檸檬",
      subtitle: "人要吃飯才能創作、網站要錢才能運作。",
      tiers: [
        { amt: 20, label: "Claude 訂閱費" },
        { amt: 60, label: "一杯飲料" },
        { amt: 120, label: "一份便當" },
      ],
    },
  },
  articles,
  donations: [
    { id: "d1", date: "2026 / 06 / 05", amt: 60,  name: "路過的水鳥", note: "新家開張賀！" },
    { id: "d2", date: "2026 / 06 / 12", amt: 120, name: "夜貓讀者",   note: "敲碗《到底》續篇" },
    { id: "d3", date: "2026 / 06 / 18", amt: 20,  name: "匿名訪客",   note: "" },
    { id: "d4", date: "2026 / 06 / 24", amt: 200, name: "MyGO 同好",  note: "ANON TOKYO 太好用" },
    { id: "d5", date: "2026 / 06 / 29", amt: 60,  name: "潛水讀者",   note: "" },
    { id: "d6", date: "2026 / 07 / 01", amt: 20,  name: "匿名訪客",   note: "加油" },
  ],
  comments: [
    { id: "c1", name: "路過的水鳥", date: "2026 / 06 / 28", text: "從巴哈小屋追過來的，新家很讚，簽到！", status: "visible", reply: "歡迎入住，記得常來。", articleId: null },
    { id: "c2", name: "夜貓讀者", date: "2026 / 06 / 25", text: "《到底》的結尾看了三遍，後勁很強。想問有沒有出版的打算？", status: "visible", reply: "謝謝！那篇結尾改了七次。出版還很遙遠，先在湖邊慢慢寫。", articleId: (articles.find(a => a.title.includes("到底")) || {}).id || null },
    { id: "c3", name: "MyGO 同好", date: "2026 / 06 / 19", text: "ANON TOKYO 每天都在用，關鍵字查梗圖太方便了，感謝大大。", status: "visible", reply: "", articleId: null },
    { id: "c4", name: "匿名訪客", date: "2026 / 06 / 30", text: "請問接不接網站案子？", status: "pending", reply: "", articleId: null },
    { id: "c5", name: "潛水讀者", date: "2026 / 07 / 01", text: "改版記事看完了，新家很有味道，期待技術筆記！", status: "visible", reply: "會努力更新的！", articleId: "a0" },
  ],
  stats: { daily: seedStats() },
};

export const Lake = {
  seed,
  load() {
    if (typeof window === "undefined") return JSON.parse(JSON.stringify(seed));
    try {
      const raw = localStorage.getItem(KEY);
      if (raw) {
        const saved = JSON.parse(raw);
        if (saved && saved.version === seed.version) return saved;
      }
    } catch (e) { /* fall through to seed */ }
    return JSON.parse(JSON.stringify(seed));
  },
  save(state) {
    if (typeof window === "undefined") return;
    try { localStorage.setItem(KEY, JSON.stringify(state)); } catch (e) {}
  },
  reset() {
    if (typeof window !== "undefined") {
      try { localStorage.removeItem(KEY); } catch (e) {}
    }
    return JSON.parse(JSON.stringify(seed));
  },
  uid() { return Math.random().toString(36).slice(2, 9); },
  today() { return "2026 / 07 / 02"; },
};
