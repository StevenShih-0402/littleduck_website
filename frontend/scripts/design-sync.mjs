// frontend/scripts/design-sync.mjs
// 設計同步工具：比對從 claude.ai 拉取的設計檔與同步基準，並在移植完成後更新基準

import { createHash } from "node:crypto";
import { existsSync, readdirSync, readFileSync, statSync, writeFileSync } from "node:fs";
import { dirname, join, relative, sep } from "node:path";
import { fileURLToPath } from "node:url";

const DEFAULT_MANIFEST = join(dirname(fileURLToPath(import.meta.url)), "..", "design-sync", "manifest.json");

function sha256(path) {
  return createHash("sha256").update(readFileSync(path)).digest("hex");
}

function walk(dir) {
  return readdirSync(dir, { withFileTypes: true }).flatMap(entry => {
    const path = join(dir, entry.name);
    return entry.isDirectory() ? walk(path) : [path];
  });
}

// include 以 / 結尾代表整個目錄，否則代表單一檔案
function isIncluded(path, include) {
  return include.some(rule => (rule.endsWith("/") ? path.startsWith(rule) : path === rule));
}

// 掃描拉取目錄，回傳納入同步範圍的 { 設計檔路徑: sha256 }
export function hashPulled(pulledDir, include) {
  if (!existsSync(pulledDir) || !statSync(pulledDir).isDirectory()) {
    throw new Error(`找不到拉取目錄：${pulledDir}`);
  }
  const hashes = {};
  for (const file of walk(pulledDir)) {
    const path = relative(pulledDir, file).split(sep).join("/");
    if (isIncluded(path, include)) hashes[path] = sha256(file);
  }
  // 空結果多半是目錄給錯，不能當成「設計檔全被移除」
  if (Object.keys(hashes).length === 0) {
    throw new Error(`拉取目錄內沒有納入同步範圍的檔案：${pulledDir}`);
  }
  return hashes;
}

// 只依內容雜湊判定差異，artifact 版本號不參與判斷
export function diff(manifest, hashes) {
  const baseline = manifest.files;
  const paths = Object.keys(hashes).sort();
  return {
    added: paths.filter(p => !baseline[p]).map(path => ({ path, targets: [] })),
    changed: paths
      .filter(p => baseline[p] && baseline[p].sha256 !== hashes[p])
      .map(path => ({ path, targets: baseline[path].targets })),
    removed: Object.keys(baseline)
      .filter(p => !hashes[p])
      .sort()
      .map(path => ({ path, targets: baseline[path].targets })),
  };
}

// 回傳新的基準，既有檔案保留 targets，新檔案的 targets 留空待補
export function buildManifest(manifest, hashes, version, date) {
  const files = {};
  for (const path of Object.keys(hashes).sort()) {
    files[path] = { sha256: hashes[path], targets: manifest.files[path]?.targets ?? [] };
  }
  return { ...manifest, syncedVersion: version, syncedAt: date, files };
}

export function formatReport(result) {
  const { added, changed, removed } = result;
  if (added.length + changed.length + removed.length === 0) return "沒有設計變更。";
  const lines = [];
  const section = (title, items, describe) => {
    if (items.length === 0) return;
    lines.push(`${title}（${items.length}）`);
    for (const item of items) lines.push(`  ${item.path}`, `    ${describe(item)}`);
  };
  section("已變更", changed, item => `→ ${item.targets.join(", ")}`);
  section("新增", added, () => "→ 尚無前端對應");
  section("移除", removed, item => `原對應 → ${item.targets.join(", ")}`);
  return lines.join("\n");
}

function parseArgs(argv) {
  const [command, pulledDir, ...rest] = argv;
  const options = {};
  for (let i = 0; i < rest.length; i += 2) {
    if (!rest[i].startsWith("--") || rest[i + 1] === undefined) throw new Error(`無法解析參數：${rest[i]}`);
    options[rest[i].slice(2)] = rest[i + 1];
  }
  return { command, pulledDir, options };
}

// 回傳 exit code：0 無變更或更新成功、1 有差異
export function run(argv, log = console.log) {
  const { command, pulledDir, options } = parseArgs(argv);
  if (!["check", "update"].includes(command) || !pulledDir) {
    throw new Error("用法：design-sync.mjs <check|update> <拉取目錄> [--version <版本>] [--manifest <路徑>]");
  }
  const manifestPath = options.manifest ?? DEFAULT_MANIFEST;
  const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
  const hashes = hashPulled(pulledDir, manifest.source.include);

  if (command === "check") {
    const result = diff(manifest, hashes);
    log(`基準版本：${manifest.syncedVersion}（${manifest.syncedAt}）`);
    log(formatReport(result));
    return result.added.length + result.changed.length + result.removed.length === 0 ? 0 : 1;
  }

  if (!options.version) throw new Error("update 需要 --version <artifact 版本>");
  const date = options.date ?? new Date().toISOString().slice(0, 10);
  const next = buildManifest(manifest, hashes, options.version, date);
  writeFileSync(manifestPath, JSON.stringify(next, null, 2) + "\n");
  log(`基準已更新為 ${options.version}。`);
  const missing = Object.keys(next.files).filter(path => next.files[path].targets.length === 0);
  if (missing.length > 0) log(`以下檔案還沒有 targets，請補上：\n  ${missing.join("\n  ")}`);
  return 0;
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  try {
    process.exitCode = run(process.argv.slice(2));
  } catch (error) {
    console.error(error.message);
    process.exitCode = 2;
  }
}
