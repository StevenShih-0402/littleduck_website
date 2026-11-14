// frontend/next.config.mjs
// Next.js 設定檔：啟用 React Strict Mode；output standalone 供 Docker 映像只帶執行所需檔案

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: "standalone",
};

export default nextConfig;
