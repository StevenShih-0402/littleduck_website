// frontend/app/layout.jsx
// 根 layout：以 next/font 載入五種字型、掛全域樣式與 Lake Provider

import "./globals.css";
import { Crimson_Pro, JetBrains_Mono, Inter, Noto_Serif_TC, Noto_Sans_TC } from "next/font/google";
import { LakeProvider } from "@/store/lake-context";

const crimson = Crimson_Pro({
  subsets: ["latin"],
  weight: ["400", "600", "700"],
  style: ["normal", "italic"],
  variable: "--font-crimson",
});
const jetbrains = JetBrains_Mono({
  subsets: ["latin"],
  weight: ["400", "500", "600", "700"],
  variable: "--font-jetbrains",
});
const inter = Inter({
  subsets: ["latin"],
  weight: ["400", "500", "600", "700", "800", "900"],
  variable: "--font-inter",
});
const notoSerifTC = Noto_Serif_TC({
  subsets: ["latin"],
  weight: ["400", "500", "600", "700", "900"],
  variable: "--font-noto-serif-tc",
});
const notoSansTC = Noto_Sans_TC({
  subsets: ["latin"],
  weight: ["400", "500", "700", "900"],
  variable: "--font-noto-sans-tc",
});

export const metadata = {
  title: "小鴨ㄉ創作湖畔 v2 · 小屋 + 後台",
  description: "個人小屋：部落格 × 作品展示 × 後台管理原型",
};

export default function RootLayout({ children }) {
  const fontVars = [crimson, jetbrains, inter, notoSerifTC, notoSansTC]
    .map((f) => f.variable)
    .join(" ");
  return (
    <html lang="zh-TW" className={fontVars} suppressHydrationWarning>
      <body className="sans">
        <LakeProvider>{children}</LakeProvider>
      </body>
    </html>
  );
}
