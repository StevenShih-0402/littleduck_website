import { test } from "node:test";
import assert from "node:assert/strict";
import {
  SPONSOR_AMOUNT_MIN, SPONSOR_AMOUNT_MAX, validateSponsorAmount, resultView, nextPollDelay,
} from "./sponsor.mjs";

test("金額常數與後端一致", () => {
  assert.equal(SPONSOR_AMOUNT_MIN, 10);
  assert.equal(SPONSOR_AMOUNT_MAX, 20000);
});

test("validateSponsorAmount：邊界值", () => {
  assert.equal(validateSponsorAmount(10).ok, true);
  assert.equal(validateSponsorAmount(20000).ok, true);
  assert.equal(validateSponsorAmount(9).ok, false);
  assert.equal(validateSponsorAmount(20001).ok, false);
  assert.equal(validateSponsorAmount(0).ok, false);
  assert.equal(validateSponsorAmount(-50).ok, false);
});

test("validateSponsorAmount：非整數與非數字", () => {
  for (const v of [10.5, NaN, Infinity, "100", null, undefined]) {
    const r = validateSponsorAmount(v);
    assert.equal(r.ok, false);
  }
  const r = validateSponsorAmount(5);
  assert.equal(r.ok, false);
  assert.match(r.message, /10/);
  assert.match(r.message, /20000/);
});

test("resultView：三種狀態與其他", () => {
  assert.equal(resultView("paid"), "paid");
  assert.equal(resultView("pending"), "pending");
  assert.equal(resultView("failed"), "failed");
  for (const v of ["", "PAID", null, undefined, 1, "unknown"]) {
    assert.equal(resultView(v), "missing");
  }
});

test("nextPollDelay：遞增且有限次數後回 null", () => {
  let n = 0;
  let prev = 0;
  while (nextPollDelay(n) !== null) {
    const d = nextPollDelay(n);
    assert.ok(d >= prev && d > 0);
    prev = d;
    n++;
    assert.ok(n < 100);
  }
  assert.ok(n >= 5);
  assert.equal(nextPollDelay(n), null);
  assert.equal(nextPollDelay(-1), null);
  assert.equal(nextPollDelay(1.5), null);
});
