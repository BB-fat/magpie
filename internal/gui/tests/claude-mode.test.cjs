const assert = require("node:assert/strict");
const fs = require("node:fs/promises");
const path = require("node:path");
const { test } = require("node:test");
const { chromium, webkit } = require("playwright");

const assets = path.resolve(__dirname, "../assets");

function fixture(lang, posts) {
  const provider = {
    id: "claude", name: "Claude", icon: "claude-color", models: [{ id: "claude-sonnet-5", on: true }],
    agents: [], fallback: [], headers: {}, keyList: [], proxy: "", claudeTuiHere: true, claudeTuiLive: true,
    claudeTuiNow: {}, accountClaudeModes: {},
    account: { agent: "claude", agentName: "Claude Code", user: "side@example.com", plan: "MAX",
      logins: [{ user: "side@example.com", plan: "MAX", on: true }] },
  };
  const providers = { providers: [provider], presets: [], excluded: [], gateway: { running: true, window: true } };
  let fail = false;
  return {
    failNext() { fail = true; },
    async route(route) {
      const req = route.request(), url = new URL(req.url());
      const json = data => route.fulfill({ json: data });
      if (url.pathname === "/boot.js") return route.fulfill({ contentType: "text/javascript", body: `window.bootPrefs=${JSON.stringify({ lang, theme: "light", web: true })};` });
      if (url.pathname === "/wails/runtime.js") return route.fulfill({ contentType: "text/javascript", body: "export const Window={};" });
      if (url.pathname === "/api/state") return json({ agents: [], profiles: [], settings: { lang, theme: "light" } });
      if (url.pathname === "/api/providers") return json(providers);
      if (url.pathname === "/api/providers/claude-mode") {
        const body = req.postDataJSON();
        posts.push(body);
        if (fail) {
          fail = false;
          return route.fulfill({ status: 400, json: { error: "Mode change failed" } });
        }
        provider.accountClaudeModes = body.mode === "tui" ? { "side@example.com": "tui" } : {};
        return json(providers);
      }
      if (url.pathname === "/api/login/usage") return json({});
      if (url.pathname === "/api/groups") return json({ groups: [] });
      if (url.pathname === "/api/plugins") return json({ plugins: [] });
      if (url.pathname === "/api/usage/quotas") return json([]);
      if (url.pathname.startsWith("/api/")) return json({});
      const file = path.join(assets, url.pathname === "/" ? "index.html" : url.pathname);
      const contentType = { ".html": "text/html", ".js": "text/javascript", ".css": "text/css", ".svg": "image/svg+xml", ".png": "image/png" }[path.extname(file)];
      return route.fulfill({ contentType, body: await fs.readFile(file) });
    },
  };
}

for (const engine of process.env.BROWSER ? [process.env.BROWSER] : ["chromium", "webkit"]) {
  for (const lang of ["en", "zh", "zh-TW", "ja", "de"]) {
    for (const width of [1100, 420]) {
      test(`${engine} ${lang} ${width}: Claude account mode persists and a failed save keeps its mode`, async t => {
        const browser = await (engine === "webkit" ? webkit.launch() : chromium.launch({ channel: "chromium" }));
        t.after(() => browser.close());
        const page = await browser.newPage({ viewport: { width, height: 800 }, reducedMotion: "reduce" });
        page.setDefaultTimeout(5000);
        const posts = [], errors = [];
        const state = fixture(lang, posts);
        page.on("pageerror", e => errors.push(e.message));
        await page.route("**/*", state.route);
        await page.goto("http://magpie.test/?view=providers");
        await page.locator(".row.provider", { hasText: "Claude" }).first().click();
        const pill = page.locator('.acc[data-account-id="side@example.com"] .amode');
        await pill.waitFor();
        assert.equal(await pill.evaluate(el => el.classList.contains("set")), false);
        await pill.click();
        const options = page.locator('.proto-menu .pm-item');
        await options.nth(1).click();
        await page.waitForFunction(() => document.querySelector('.acc .amode')?.classList.contains("pending"));
        assert.deepEqual(posts[0], { id: "claude", account: "side@example.com", mode: "tui" });
        assert.equal(await pill.evaluate(el => el.classList.contains("set")), true);
        await page.reload();
        if (!(await pill.isVisible())) await page.locator(".row.provider", { hasText: "Claude" }).first().click();
        await pill.waitFor();
        assert.equal(await pill.evaluate(el => el.classList.contains("pending")), true);
        state.failNext();
        await pill.click();
        await page.locator('.proto-menu .pm-item').first().click();
        await page.getByText("Mode change failed", { exact: true }).waitFor();
        assert.equal(await pill.evaluate(el => el.classList.contains("set")), true);
        await pill.click();
        await page.locator('.proto-menu .pm-item').first().click();
        await page.waitForFunction(() => !document.querySelector('.acc .amode')?.classList.contains("set"));
        assert.equal(await pill.evaluate(el => el.classList.contains("pending")), false);
        assert.deepEqual(posts.at(-1), { id: "claude", account: "side@example.com", mode: "standard" });
        assert.deepEqual(errors, []);
      });
    }
  }
}
