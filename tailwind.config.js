// frontend/tailwind.config.js
// Tailwind CSS 設定：掃描 app、components、store 內的 JSX 類名

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./app/**/*.{js,jsx}",
    "./components/**/*.{js,jsx}",
    "./store/**/*.{js,jsx}",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
};
