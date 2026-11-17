import { test } from "node:test";
import assert from "node:assert/strict";
import { FEATURES, isFeatureEnabled, gateMode } from "./feature-flags.mjs";

test("FEATURES 含四個開關 key", () => {
  assert.deepEqual(Object.values(FEATURES).sort(), ["article_comments", "guestbook", "sponsor", "sponsor_payment"]);
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

test("gateMode：sponsor 與 sponsor_payment 兩個開關組合 × 是否管理員", () => {
  const cases = [
    // [sponsor, sponsor_payment, 預期外層, 預期內層]
    [true, true, "show", "show"],
    [true, false, "show", "hide"],
    [false, true, "hide", "show"],
    [false, false, "hide", "hide"],
  ];
  for (const [a, b, outer, inner] of cases) {
    const flags = { sponsor: a, sponsor_payment: b };
    // 一般訪客：任一層 hide 就看不到
    assert.equal(gateMode(flags, FEATURES.SPONSOR, false), outer);
    assert.equal(gateMode(flags, FEATURES.SPONSOR_PAYMENT, false), inner);
    // 管理員：停用的那層為 preview
    assert.equal(gateMode(flags, FEATURES.SPONSOR, true), a ? "show" : "preview");
    assert.equal(gateMode(flags, FEATURES.SPONSOR_PAYMENT, true), b ? "show" : "preview");
  }
});
