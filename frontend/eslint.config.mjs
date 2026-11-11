// frontend/eslint.config.mjs
// ESLint flat config：採 Next.js 官方 core-web-vitals 規則組（含 Next、React、React Hooks 規則）
// 既有違規只能以帶理由的單行註解豁免，不在這裡關閉或降級規則（見 openspec frontend-project-structure 規格）

import { defineConfig, globalIgnores } from 'eslint/config';
import nextVitals from 'eslint-config-next/core-web-vitals';

const eslintConfig = defineConfig([
  ...nextVitals,
  // 覆寫 eslint-config-next 的預設忽略清單（內容與預設相同，只排除建置產物）
  globalIgnores(['.next/**', 'out/**', 'build/**', 'next-env.d.ts']),
]);

export default eslintConfig;
