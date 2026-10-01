// Browser verification runs against a temporary fixture, never service actions.
const fs = require('node:fs');
const http = require('node:http');
const path = require('node:path');
const assert = require('node:assert/strict');
const { chromium } = require('/check/node_modules/playwright');
const root = '/work';
const fixture = JSON.parse(fs.readFileSync('/check/status.json', 'utf8'));
const source = fs.readFileSync(path.join(root, 'vohive.js'), 'utf8');
new Function(source); // LuCI view modules intentionally have a top-level return.
const harness = `<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><link rel="stylesheet" href="/luci/argon/css/cascade.css"><style>
body { margin:0; background:#f4f5f8; } .preview-sidebar{position:fixed;inset:0 auto 0 0;width:215px;background:#fff;padding:24px 14px;border-right:1px solid #e8eaf0;color:#4f5970;font:15px sans-serif}.preview-logo{font-size:30px;font-weight:bold;color:#606fe8;margin:8px 22px 32px}.preview-nav{padding:14px 18px;border-radius:9px;margin:3px 0}.preview-nav.active{background:#6264e8;color:white}.preview-subnav{padding:14px 22px 14px 38px;font-size:13px}.preview-subnav.active{color:#6264e8;background:#f3f3fd}.preview-main{margin-left:215px;padding:32px}.preview-breadcrumb{font:12px sans-serif;color:#8590a4;margin:0 0 18px} @media(max-width:700px){.preview-sidebar{display:none}.preview-main{margin-left:0;padding:14px}.preview-breadcrumb{margin:4px 0 14px}}
</style></head><body><aside class="preview-sidebar"><div class="preview-logo">iStoreOS</div><div class="preview-nav">⌂　首页</div><div class="preview-nav">◎　网络向导</div><div class="preview-nav">▦　状态</div><div class="preview-nav">⚙　系统</div><div class="preview-nav">▧　iStore</div><div class="preview-nav active">⚙　服务</div><div class="preview-subnav">PassWall</div><div class="preview-subnav">应用过滤</div><div class="preview-subnav">易有云文件管理器</div><div class="preview-subnav">Tailscale</div><div class="preview-subnav">动态 DNS</div><div class="preview-subnav active">VoHive Plus</div><div class="preview-subnav">网络唤醒</div></aside><main class="preview-main"><div class="preview-breadcrumb">服务　/　VoHive Plus</div><div id="app"></div></main><script>
function E(tag, attrs, children) {
    var el = document.createElement(tag);
    Object.keys(attrs || {}).forEach(function(key) {
        var value = attrs[key];
        if (typeof value === 'function') el.addEventListener(key, value);
        else if (value === true) el.setAttribute(key, '');
        else if (value !== false && value !== null && value !== undefined) el.setAttribute(key, String(value));
    });
    function append(value) {
        if (Array.isArray(value)) value.forEach(append);
        else if (value instanceof Node) el.appendChild(value);
        else if (value !== undefined && value !== null) el.appendChild(document.createTextNode(String(value)));
    }
    append(children); return el;
}
window.__notifications = [];
window.__actions = [];
window.__polls = [];
window.__rpcResult = 0;
window.__readonly = false;
window.__fail = false;
window.__fixture = null;
window.__copied = '';
Object.defineProperty(navigator, 'clipboard', { value: undefined, configurable: true });
document.execCommand = function(action) { if (action === 'copy') { window.__copied = document.querySelector('textarea').value; return true; } return false; };
var L = { hasViewPermission: function() { return !window.__readonly; }, resource: function(p) { return '/assets/' + p.split('/').pop(); }, bind: function(fn, scope) { return fn.bind(scope); } };
var view = { extend: function(v) { return v; } };
var rpc = { declare: function(spec) { return function(name, action) { if (spec.method === 'init') { window.__actions.push(action); return Promise.resolve(window.__rpcResult); } return window.__fail ? Promise.reject(new Error('network failed')) : Promise.resolve({ vohive: { enabled: true, running: true } }); }; } };
var fileApi = { exec: function() { return window.__fail ? Promise.reject(new Error('network failed')) : Promise.resolve({code:0, stdout:JSON.stringify(window.__fixture)}); } };
var ui = { createHandlerFn: function(scope, name, arg) { return function() { return scope[name](arg); }; }, addNotification: function(title, node, type) { window.__notifications.push({ text:node.textContent, type:type }); } };
var poll = { add: function(fn) { window.__polls.push(fn); }, remove: function(fn) { window.__polls = window.__polls.filter(function(p) { return p !== fn; }); } };
Promise.all([fetch('/source').then(function(r) { return r.text(); }),fetch('/fixture').then(function(r) { return r.json(); })]).then(function(data) {
    window.__fixture = data[1];
    window.__view = new Function('view','rpc','fs','ui','poll','L','E', data[0])(view,rpc,fileApi,ui,poll,L,E);
    document.getElementById('app').appendChild(window.__view.render({service:{enabled:true,running:true},info:window.__fixture}));
    window.__ready = true;
});
</script></body></html>`;

const server = http.createServer((req, res) => {
    const url = new URL(req.url, 'http://localhost');
    if (url.pathname === '/') { res.setHeader('Content-Type', 'text/html; charset=utf-8'); return res.end(harness); }
    if (url.pathname === '/fixture') { res.setHeader('Content-Type', 'application/json'); return res.end(JSON.stringify(fixture)); }
    if (url.pathname === '/source') { res.setHeader('Content-Type', 'text/plain'); return res.end(source); }
    let file;
    if (url.pathname === '/assets/vohive.css') file = path.join(root, 'vohive.css');
    if (url.pathname.startsWith('/luci/')) {
        const resolved = path.resolve('/luci', url.pathname.slice(6));
        if (resolved.startsWith('/luci/')) file = resolved;
    }
    if (file && fs.existsSync(file)) {
        res.setHeader('Content-Type', file.endsWith('.css') ? 'text/css' : 'application/octet-stream');
        return fs.createReadStream(file).pipe(res);
    }
    res.writeHead(404); res.end();
});

(async () => {
    await new Promise(resolve => server.listen(8769, '127.0.0.1', resolve));
    const browser = await chromium.launch({ executablePath: '/usr/bin/chromium', args: ['--no-sandbox'] });
    const page = await browser.newPage({ viewport: { width: 1600, height: 1100 }, locale: 'zh-CN', timezoneId: 'Asia/Shanghai' });
    const errors = [];
    page.on('pageerror', error => errors.push(error.message));
    await page.goto('http://127.0.0.1:8769/');
    await page.waitForFunction(() => window.__ready);
    await page.locator('.vh-page').screenshot({ path: '/check/preview-desktop.png' });
    assert.equal(await page.locator('.vh-heading').textContent(), 'VoHive');
    assert.equal(await page.locator('.vh-badge').textContent(), '运行中');
    assert.equal(await page.locator('button:has-text("启动")').isDisabled(), true);
    assert.equal(await page.locator('button:has-text("停止")').isDisabled(), false);
    assert.equal(await page.locator('.vh-log-body').textContent().then(s => /Worker|active_workers|\[GIN\]|\x1b/.test(s)), false);
    assert.equal(await page.locator('.vh-log-row').count(), fixture.logs.filter(l => !l.heartbeat).length);
    console.log('通过：真实状态、中文日志、默认心跳筛选与按钮状态');

    await page.getByLabel('日志等级', { exact: true }).selectOption('error');
    assert.equal(await page.locator('.vh-log-row').count(), fixture.logs.filter(l => l.level === 'error' && !l.heartbeat).length);
    await page.getByLabel('日志等级', { exact: true }).selectOption('');
    await page.getByLabel('搜索日志', { exact: true }).fill('不存在的测试关键词');
    assert.equal(await page.locator('.vh-empty').count(), 1);
    await page.getByLabel('搜索日志', { exact: true }).fill('');
    await page.getByText('隐藏心跳', { exact: true }).click();
    assert.equal(await page.locator('.vh-log-row').count(), fixture.logs.length);
    await page.getByRole('button', { name: '复制日志' }).click();
    assert(await page.evaluate(() => __copied.includes('[信息]') && !__copied.includes('[GIN]')));
    console.log('通过：日志等级、搜索、心跳开关与 HTTP 页面复制');

    await page.evaluate(() => { __fail = true; return __view.refreshStatus(); });
    assert.equal(await page.locator('.vh-badge').textContent(), '运行中');
    assert.equal(await page.locator('.vh-notice').isVisible(), true);
    await page.evaluate(() => { __fail = false; return __view.refreshStatus(); });
    assert.equal(await page.locator('.vh-notice').isVisible(), false);
    await page.evaluate(() => { __view.autoRefresh.checked = false; __fixture.running = false; return __polls[0](); });
    assert.equal(await page.locator('.vh-badge').textContent(), '运行中');
    await page.evaluate(() => { __view.autoRefresh.checked = true; return __polls[0](); });
    assert.equal(await page.locator('.vh-badge').textContent(), '已停止');
    assert.equal(await page.locator('button:has-text("启动")').isDisabled(), false);
    assert.equal(await page.locator('button:has-text("停止")').isDisabled(), true);
    console.log('通过：断线保留数据、恢复刷新、暂停轮询和停止状态');

    await page.evaluate(() => __view.runAction('start'));
    assert.equal(await page.evaluate(() => __actions.at(-1)), 'start');
    assert.equal(await page.evaluate(() => __notifications.at(-1).type), 'info');
    await page.evaluate(() => { __rpcResult = 1; return __view.runAction('start'); });
    assert.equal(await page.evaluate(() => __notifications.at(-1).type), 'error');
    await page.evaluate(() => { __readonly = true; const data = {service:{enabled:true},info:__fixture}; document.getElementById('app').replaceChildren(__view.render(data)); });
    assert.equal(await page.locator('button:has-text("启动")').isDisabled(), true);
    const actionCount = await page.evaluate(() => __actions.length);
    await page.evaluate(() => __view.runAction('start'));
    assert.equal(await page.evaluate(() => __actions.length), actionCount);
    console.log('通过：操作成功与失败提示、只读权限（模拟操作）');

    await page.evaluate(() => {
        __readonly = false; __fixture.running = true;
        __fixture.logs.push({ time: '2026-10-01 17:30:00', level: 'error', message: '<img src=x onerror="window.__xss=true">设备错误', detail: '', heartbeat: false });
        document.getElementById('app').replaceChildren(__view.render({service:{enabled:true},info:__fixture}));
    });
    assert.equal(await page.locator('.vh-log-message img').count(), 0);
    assert.equal(await page.evaluate(() => Boolean(window.__xss)), false);
    await page.evaluate(() => { __fixture.logs.pop(); __view.updateStatus({service:{enabled:true},info:__fixture}); });
    console.log('通过：日志内容按文字显示，避免 HTML 注入');

    await page.setViewportSize({ width: 390, height: 844 });
    assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth), false);
    await page.locator('.vh-page').screenshot({ path: '/check/preview-mobile.png' });
    await page.getByText('服务诊断', { exact: true }).click();
    assert.equal(await page.locator('.vh-diagnostic-content').isVisible(), true);
    assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth), false);
    console.log('通过：手机布局、诊断展开和长路径无横向溢出');
    await page.setViewportSize({ width: 1600, height: 1100 });
    await page.evaluate(() => { document.querySelector('.vh-page').setAttribute('data-theme','dark'); document.querySelector('details').open = false; });
    await page.locator('.vh-page').screenshot({ path: '/check/preview-dark.png' });
    assert.deepEqual(errors, []);
    console.log('通过：深色样式与浏览器无脚本异常');
    await browser.close();
    server.close();
})().catch(error => { console.error(error); server.close(); process.exit(1); });
