import { test } from "node:test";
import assert from "node:assert/strict";
import { FEATURES, isFeatureEnabled, gateMode } from "./feature-flags.mjs";

test("FEATURES 含三個開關 key", () => {
  assert.deepEqual(Object.values(FEATURES).sort(), ["article_comments", "guestbook", "sponsor"]);
});

test("isFeatureEnabled 只認嚴格 true", () => {
  assert.equal(isFeatureEnabled({ guestbook: true }, "guestbook"), true);
  assert.equal(isFeatureEnabled({ guestbook: false }, "guestbook"), false);
  assert.equal(isFeatureEnabled({}, "guestbook"), false);
  for (const v of ["true", 1, "yes", null, undefined, {}, []]) {
    assert.equal(isFeatureEnabled({ guestbook: v }, "guestbook"), false);
  }
});

test("isFeatureEnabled 在 flags 為 null 或 undefined 時為 false", () => {
  assert.equal(isFeatureEnabled(null, "guestbook"), false);
  assert.equal(isFeatureEnabled(undefined, "guestbook"), false);
});

test("gateMode：啟用一律 show，不分是否預覽", () => {
  for (const p of [true, false]) {
    assert.equal(gateMode({ sponsor: true }, "sponsor", p), "show");
  }
});

test("gateMode：停用時管理員 preview、訪客 hide", () => {
  assert.equal(gateMode({ sponsor: false }, "sponsor", true), "preview");
  assert.equal(gateMode({ sponsor: false }, "sponsor", false), "hide");
});

test("gateMode：缺 key、非布林值、flags 為 null 視同停用", () => {
  for (const flags of [{}, { sponsor: "true" }, { sponsor: 1 }, null]) {
    assert.equal(gateMode(flags, "sponsor", true), "preview");
    assert.equal(gateMode(flags, "sponsor", false), "hide");
  }
});
