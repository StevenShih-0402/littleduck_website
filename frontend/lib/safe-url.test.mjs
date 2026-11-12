import { test } from "node:test";
import assert from "node:assert/strict";
import { classifyHref } from "./safe-url.mjs";

const kinds = (list) => list.map((v) => classifyHref(v).kind);

test("站內路徑為 internal", () => {
  assert.deepEqual(kinds(["/", "/blog/a1"]), ["internal", "internal"]);
  assert.equal(classifyHref("/blog/a1").href, "/blog/a1");
});

test("http/https/mailto 為 external，不分大小寫並去除前後空白", () => {
  assert.deepEqual(kinds(["https://x", "HTTP://X", "mailto:a@b", " https://x "]), ["external", "external", "external", "external"]);
  assert.equal(classifyHref(" https://x ").href, "https://x");
});

test("空、#、#section 為 none", () => {
  assert.deepEqual(kinds(["", "#", "#section", "   ", undefined, null]), Array(6).fill("none"));
  assert.equal(classifyHref("#").href, "");
});

test("不安全或不支援的網址為 none 且不帶 href", () => {
  const bad = ["javascript:alert(1)", "JavaScript:x", "data:text/html,x", "//evil.com", "/\\evil.com", "foo/bar", "ftp://x", " javascript:x"];
  for (const v of bad) assert.deepEqual(classifyHref(v), { kind: "none", href: "" }, v);
});
