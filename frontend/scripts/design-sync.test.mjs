// frontend/scripts/design-sync.test.mjs
// design-sync 的單元測試：差異分類、基準更新與錯誤輸入

import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { afterEach, beforeEach, describe, it } from "node:test";

import { run } from "./design-sync.mjs";

const hash = text => createHash("sha256").update(text).digest("hex");

describe("design-sync", () => {
  let root, pulled, manifestPath, logs;

  const write = (path, text) => {
    const file = join(pulled, path);
    mkdirSync(dirname(file), { recursive: true });
    writeFileSync(file, text);
  };
  const exec = (...args) => run([...args, "--manifest", manifestPath], line => logs.push(line));
  const readManifest = () => JSON.parse(readFileSync(manifestPath, "utf8"));
  const output = () => logs.join("\n");

  beforeEach(() => {
    root = mkdtempSync(join(tmpdir(), "design-sync-"));
    pulled = join(root, "pulled");
    manifestPath = join(root, "manifest.json");
    logs = [];
    write("project/v2/ui.jsx", "ui");
    write("project/data.js", "data");
    write("project/compat.json", "不在同步範圍");
    write("artifact-type/app.js", "編輯器檔案");
    writeFileSync(manifestPath, JSON.stringify({
      source: { artifactUrl: "https://example.test/a", include: ["project/v2/", "project/data.js"] },
      syncedVersion: "v1",
      syncedAt: "2026-11-05",
      files: {
        "project/data.js": { sha256: hash("data"), targets: ["data/site.js"] },
        "project/v2/ui.jsx": { sha256: hash("ui"), targets: ["components/ui.jsx"] },
      },
    }));
  });

  afterEach(() => rmSync(root, { recursive: true, force: true }));

  it("內容相同時回報無變更，範圍外的檔案不影響結果", () => {
    assert.equal(exec("check", pulled), 0);
    assert.match(output(), /沒有設計變更/);
  });

  it("版本號不同但內容相同時回報無變更", () => {
    assert.equal(exec("check", pulled, "--version", "v2"), 0);
    assert.match(output(), /沒有設計變更/);
  });

  it("內容變更時列出對應的前端檔案", () => {
    write("project/v2/ui.jsx", "ui 已修改");
    assert.equal(exec("check", pulled), 1);
    assert.match(output(), /已變更（1）/);
    assert.match(output(), /project\/v2\/ui\.jsx\n\s+→ components\/ui\.jsx/);
  });

  it("新增檔案時標示尚無前端對應", () => {
    write("project/v2/new-page.jsx", "new");
    assert.equal(exec("check", pulled), 1);
    assert.match(output(), /新增（1）/);
    assert.match(output(), /project\/v2\/new-page\.jsx\n\s+→ 尚無前端對應/);
  });

  it("移除檔案時列出原本對應的前端檔案", () => {
    rmSync(join(pulled, "project/data.js"));
    assert.equal(exec("check", pulled), 1);
    assert.match(output(), /移除（1）/);
    assert.match(output(), /project\/data\.js\n\s+原對應 → data\/site\.js/);
  });

  it("check 不會改動基準", () => {
    write("project/v2/ui.jsx", "ui 已修改");
    const before = readFileSync(manifestPath, "utf8");
    exec("check", pulled);
    assert.equal(readFileSync(manifestPath, "utf8"), before);
  });

  it("update 後再 check 回報無變更，並保留 targets", () => {
    write("project/v2/ui.jsx", "ui 已修改");
    assert.equal(exec("update", pulled, "--version", "v2", "--date", "2026-12-01"), 0);
    const manifest = readManifest();
    assert.equal(manifest.syncedVersion, "v2");
    assert.equal(manifest.syncedAt, "2026-12-01");
    assert.equal(manifest.files["project/v2/ui.jsx"].sha256, hash("ui 已修改"));
    assert.deepEqual(manifest.files["project/v2/ui.jsx"].targets, ["components/ui.jsx"]);
    assert.equal(manifest.source.artifactUrl, "https://example.test/a");
    assert.equal(exec("check", pulled), 0);
  });

  it("update 移除已不存在的檔案，並提示新檔案要補 targets", () => {
    rmSync(join(pulled, "project/data.js"));
    write("project/v2/new-page.jsx", "new");
    exec("update", pulled, "--version", "v2");
    const { files } = readManifest();
    assert.equal(files["project/data.js"], undefined);
    assert.deepEqual(files["project/v2/new-page.jsx"].targets, []);
    assert.match(output(), /還沒有 targets[\s\S]*new-page\.jsx/);
  });

  it("update 缺少版本時失敗且不改動基準", () => {
    const before = readFileSync(manifestPath, "utf8");
    assert.throws(() => exec("update", pulled), /--version/);
    assert.equal(readFileSync(manifestPath, "utf8"), before);
  });

  it("拉取目錄不存在或沒有範圍內檔案時失敗，不回報成全部移除", () => {
    assert.throws(() => exec("check", join(root, "missing")), /找不到拉取目錄/);
    const empty = join(root, "empty");
    mkdirSync(empty);
    assert.throws(() => exec("check", empty), /沒有納入同步範圍/);
  });

  // ── 刻意分歧（SDP-103 store/lake.js 不再追隨 project/v2/store.js） ──
  describe("刻意分歧（divergent）", () => {
    const markDivergent = () => {
      const m = readManifest();
      m.files["project/v2/ui.jsx"].divergent = true;
      m.files["project/v2/ui.jsx"].divergentReason = "改接後端 API，刻意不追隨設計";
      m.files["project/v2/ui.jsx"].divergentSince = "2026-11-10";
      writeFileSync(manifestPath, JSON.stringify(m));
    };

    it("被標記 divergent 的設計檔變更不列為待同步，而標示為刻意分歧", () => {
      markDivergent();
      write("project/v2/ui.jsx", "ui 已修改");
      // 沒有其他待同步變更，exit 0
      assert.equal(exec("check", pulled), 0);
      assert.match(output(), /沒有設計變更/);
      assert.match(output(), /刻意分歧（不同步）（1）/);
      assert.match(output(), /project\/v2\/ui\.jsx/);
      assert.doesNotMatch(output(), /已變更（1）/);
    });

    it("其餘未標記的對應照常列為待同步", () => {
      markDivergent();
      write("project/v2/ui.jsx", "ui 已修改");   // divergent，不計
      write("project/data.js", "data 已修改");    // 一般，應列為已變更
      assert.equal(exec("check", pulled), 1);
      assert.match(output(), /已變更（1）/);
      assert.match(output(), /project\/data\.js/);
      assert.match(output(), /刻意分歧（不同步）（1）/);
    });

    it("update 不以設計檔覆寫 divergent 對應，保留原雜湊與標記", () => {
      markDivergent();
      write("project/v2/ui.jsx", "ui 已修改");
      const originalSha = hash("ui");
      assert.equal(exec("update", pulled, "--version", "v2", "--date", "2026-12-01"), 0);
      const manifest = readManifest();
      // 原雜湊保留（未被新內容雜湊覆寫），標記仍在
      assert.equal(manifest.files["project/v2/ui.jsx"].sha256, originalSha);
      assert.equal(manifest.files["project/v2/ui.jsx"].divergent, true);
      assert.equal(manifest.files["project/v2/ui.jsx"].divergentReason, "改接後端 API，刻意不追隨設計");
    });
  });
});
