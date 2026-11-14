// frontend/lib/runtime-config.test.mjs
import { test } from "node:test";
import assert from "node:assert/strict";
import { normalizeApiBaseUrl, renderRuntimeConfigScript, resolveApiBaseUrl } from "./runtime-config.mjs";

test("normalize：去結尾斜線與前後空白", () => {
  assert.equal(normalizeApiBaseUrl("https://api.example.com/"), "https://api.example.com");
  assert.equal(normalizeApiBaseUrl("https://api.example.com///"), "https://api.example.com");
  assert.equal(normalizeApiBaseUrl("  http://localhost:8080  "), "http://localhost:8080");
});

test("normalize：接受 http 與 https", () => {
  assert.equal(normalizeApiBaseUrl("http://a.test"), "http://a.test");
  assert.equal(normalizeApiBaseUrl("https://a.test"), "https://a.test");
});

test("normalize：非法值回 null", () => {
  for (const v of ["javascript:alert(1)", "/api", "api.example.com", "", "   ", "ftp://a.test", undefined, null, 5]) {
    assert.equal(normalizeApiBaseUrl(v), null, String(v));
  }
});

test("render：輸出賦值語句", () => {
  assert.equal(
    renderRuntimeConfigScript("https://a.test"),
    'window.__LD_RUNTIME_CONFIG__={"apiBaseUrl":"https://a.test"};',
  );
  assert.equal(renderRuntimeConfigScript(null), 'window.__LD_RUNTIME_CONFIG__={"apiBaseUrl":null};');
});

test("render：跳脫 < > & U+2028 U+2029", () => {
  const out = renderRuntimeConfigScript("http://a.test/</script>&x  ");
  for (const ch of ["<", ">", "&", " ", " "]) assert.ok(!out.includes(ch), "含 " + JSON.stringify(ch));
  assert.ok(out.includes("\\u003c") && out.includes("\\u003e") && out.includes("\\u0026"));
  assert.ok(out.includes("\\u2028") && out.includes("\\u2029"));
});

test("resolve：三層 fallback", () => {
  assert.equal(resolveApiBaseUrl("https://rt.test/", "https://build.test"), "https://rt.test");
  assert.equal(resolveApiBaseUrl(undefined, "https://build.test/"), "https://build.test");
  assert.equal(resolveApiBaseUrl("javascript:x", "https://build.test"), "https://build.test");
  assert.equal(resolveApiBaseUrl("", ""), "http://localhost:8080");
  assert.equal(resolveApiBaseUrl(undefined, undefined), "http://localhost:8080");
});
