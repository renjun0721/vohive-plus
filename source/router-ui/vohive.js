'use strict';
'require view';
'require rpc';
'require fs';
'require ui';
'require poll';

var callRcList = rpc.declare({ object: 'rc', method: 'list', params: [ 'name' ], expect: { '': {} } });
var callRcInit = rpc.declare({ object: 'rc', method: 'init', params: [ 'name', 'action' ] });
var levelNames = { info: '信息', warn: '警告', error: '错误', debug: '调试' };
var actionNames = { start: '启动', restart: '重启', stop: '停止' };
var iconPaths = {
    hive: ['M12 3 4 7.5v9L12 21l8-4.5v-9L12 3Z', 'm4 7.5 8 4.5 8-4.5M12 12v9', 'm8 5.3 8 4.5v4.5'],
    power: ['M12 3v9', 'M6.3 5.8a8 8 0 1 0 11.4 0'],
    refresh: ['M20 7v5h-5', 'M4 17v-5h5', 'M6.1 6a8 8 0 0 1 13.4 3M4.5 15A8 8 0 0 0 17.9 18'],
    stop: ['M6 6h12v12H6Z'],
    external: ['M14 3h7v7', 'm21 3-10 10', 'M10 3H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-5'],
    clock: ['M12 8v4l3 2', 'M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z'],
    check: ['m6 12 4 4 8-8', 'M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z'],
    port: ['M4 4h16v7H4Z', 'M4 15h16v5H4Z', 'M8 7h.01M8 17.5h.01M12 11v4'],
    logs: ['M8 3h11a2 2 0 0 1 2 2v16H8a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2Z', 'M3 7v12a2 2 0 0 0 2 2M10 8h7M10 12h7M10 16h4'],
    search: ['M18 10a7 7 0 1 1-14 0 7 7 0 0 1 14 0Z', 'm15 15 6 6'],
    copy: ['M8 8h12v12H8Z', 'M16 8V4H4v12h4'],
    chevron: ['m8 10 4 4 4-4'],
    shield: ['m12 3 8 3v6c0 5-8 9-8 9s-8-4-8-9V6l8-3Z', 'm8 12 3 3 5-6']
};

function icon(name) {
    var svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    svg.setAttribute('viewBox', '0 0 24 24');
    svg.setAttribute('class', 'vh-icon');
    svg.setAttribute('fill', 'none');
    svg.setAttribute('stroke', 'currentColor');
    svg.setAttribute('stroke-width', '1.7');
    svg.setAttribute('stroke-linecap', 'round');
    svg.setAttribute('stroke-linejoin', 'round');
    svg.setAttribute('aria-hidden', 'true');
    (iconPaths[name] || iconPaths.logs).forEach(function(d) {
        var path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        path.setAttribute('d', d);
        svg.appendChild(path);
    });
    return svg;
}

function elapsed(start, now) {
    var seconds = Math.max(0, Math.floor((now * 1000 - Date.parse(start)) / 1000));
    if (!isFinite(seconds)) return '—';
    var days = Math.floor(seconds / 86400), hours = Math.floor(seconds % 86400 / 3600), minutes = Math.floor(seconds % 3600 / 60);
    if (days) return days + ' 天 ' + hours + ' 小时';
    if (hours) return hours + ' 小时 ' + minutes + ' 分钟';
    return minutes ? minutes + ' 分钟' : seconds + ' 秒';
}

function timeText(epoch) {
    return new Date(epoch * 1000).toLocaleTimeString('zh-CN', { hour12: false });
}

function localManagementUrl(hostname) {
    var host = hostname;
    if (host.indexOf(':') >= 0 && host.charAt(0) !== '[') host = '[' + host + ']';
    return 'http://' + host + ':7575/';
}

return view.extend({
    load: function() {
        return this.fetchStatus().catch(function() { return null; });
    },

    fetchStatus: function() {
        return Promise.all([ callRcList('vohive'), fs.exec('/usr/libexec/vohive-luci-info') ]).then(function(data) {
            if (data[1].code !== 0 || !data[1].stdout) throw new Error('读取失败');
            var info = JSON.parse(data[1].stdout);
            if (!Array.isArray(info.logs) || !Array.isArray(info.ports) || !Array.isArray(info.files)) throw new Error('状态格式错误');
            return { service: data[0].vohive || {}, info: info };
        });
    },

    refreshStatus: function() {
        if (this.refreshing) return this.refreshing;
        this.refreshButton.disabled = true;
        this.refreshing = this.fetchStatus().then(L.bind(function(data) {
            this.updateStatus(data);
        }, this)).catch(L.bind(function() {
            this.notice.hidden = false;
            this.notice.textContent = '暂时无法读取服务状态，请检查连接或登录状态后重试。' + (this.snapshot ? '当前显示上次成功读取的数据。' : '');
        }, this)).finally(L.bind(function() {
            this.refreshing = null;
            this.refreshButton.disabled = false;
        }, this));
        return this.refreshing;
    },

    updateButtons: function() {
        var known = this.snapshot && this.snapshot.info.available;
        var running = known && this.snapshot.info.running;
        var transitioning = known && this.snapshot.info.restarting;
        Object.keys(this.actionButtons).forEach(L.bind(function(action) {
            this.actionButtons[action].disabled = this.readonly || this.busy || !known || transitioning || (action === 'start' ? running : !running);
        }, this));
        var label = this.busy ? '正在' + actionNames[this.pendingAction] + '服务，请稍候…' : this.readonly ? '当前为只读模式' : !known ? '等待服务状态' : transitioning ? '服务正在重启' : running ? '停止服务' : '启动服务';
        this.powerButton.disabled = this.readonly || this.busy || !known || transitioning;
        this.powerButton.title = label;
        this.powerButton.setAttribute('aria-label', label);
        this.powerButton.setAttribute('aria-busy', String(this.busy));
        this.powerButton.setAttribute('aria-pressed', String(Boolean(running)));
        this.powerButton.classList.toggle('green', Boolean(running));
        if (this.busy) this.cards.status.note.textContent = label;
        else if (this.snapshot) {
            var info = this.snapshot.info;
            var health = { healthy: '健康检查正常', unhealthy: '健康检查异常', starting: '正在进行健康检查', unknown: '未配置健康检查' };
            this.cards.status.note.textContent = !known ? '暂时无法读取容器信息' : running ? (health[info.health] || '健康状态未知') : '服务尚未运行';
        }
    },

    toggleService: function() {
        if (this.powerButton.disabled) return Promise.resolve();
        return this.runAction(this.snapshot.info.running ? 'stop' : 'start');
    },

    runAction: function(action) {
        if (this.readonly || this.busy || !actionNames[action]) return Promise.resolve();
        this.busy = true;
        this.pendingAction = action;
        this.updateButtons();
        return callRcInit('vohive', action).then(L.bind(function(result) {
            // rc.init returns an exit code, rather than a successful boolean.
            if (result !== 0) throw new Error('操作失败');
            ui.addNotification(null, E('p', {}, actionNames[action] + '指令已执行，服务状态将在刷新后更新。'), 'info');
            return this.refreshStatus();
        }, this)).catch(function() {
            ui.addNotification(null, E('p', {}, '服务操作失败，请检查管理权限和服务启动脚本。'), 'error');
        }).finally(L.bind(function() {
            this.busy = false;
            this.pendingAction = null;
            this.updateButtons();
        }, this));
    },

    updateStatus: function(data) {
        this.snapshot = data;
        var info = data.info, known = info.available;
        var state = !known ? '未知' : info.restarting ? '重启中' : info.paused ? '已暂停' : info.running ? '运行中' : '已停止';
        this.badge.className = 'vh-badge ' + (!known ? '' : info.restarting || info.paused ? 'warning' : info.running ? 'running' : 'stopped');
        this.badgeLabel.textContent = state;
        this.cards.status.value.textContent = state;
        this.cards.status.value.style.color = known && info.running && !info.paused ? 'var(--vh-green)' : 'var(--vh-text)';
        this.cards.boot.value.textContent = typeof data.service.enabled === 'boolean' ? (data.service.enabled ? '已开启' : '已关闭') : '未知';
        this.cards.boot.note.textContent = data.service.enabled ? '设备开机后自动启动' : '可在系统启动项中设置';
        var webListening = info.ports.some(function(port) { return port.port === 7575; });
        this.cards.port.value.textContent = '7575';
        this.cards.port.note.textContent = webListening ? '管理页面端口正在监听' : '管理页面端口未监听';
        this.cards.uptime.value.textContent = known && info.running ? elapsed(info.started_at, info.checked_at) : '—';
        this.cards.uptime.note.textContent = known && info.running ? '本次服务持续运行时间' : '启动后显示运行时间';
        this.updated.textContent = '更新于 ' + timeText(info.checked_at);
        this.logNote.textContent = '最近 ' + info.logs.length + ' 条记录 · 最新记录在前';
        this.notice.hidden = known;
        this.notice.textContent = known ? '' : '暂时无法读取容器信息，请检查 Docker 服务和 VoHive Plus 容器。';
        this.updateButtons();
        this.renderLogs();
        this.renderDiagnostics();
    },

    filteredLogs: function() {
        var logs = this.snapshot ? this.snapshot.info.logs : [];
        var query = this.searchInput.value.trim().toLowerCase(), level = this.levelSelect.value;
        return logs.filter(L.bind(function(log) {
            return (!level || log.level === level) && (!this.hideHeartbeat.checked || !log.heartbeat) && (!query || [log.time, log.message, log.detail, levelNames[log.level]].join(' ').toLowerCase().indexOf(query) >= 0);
        }, this)).slice().reverse();
    },

    renderLogs: function() {
        var logs = this.filteredLogs(), total = this.snapshot ? this.snapshot.info.logs.length : 0;
        this.logCount.textContent = '显示 ' + logs.length + ' / ' + total + ' 条记录';
        this.copyButton.disabled = !logs.length;
        var signature = JSON.stringify(logs);
        if (signature === this.logSignature) return;
        this.logSignature = signature;
        var scrollTop = this.logBody.scrollTop;
        var nodes = logs.length ? logs.map(function(log) {
            return E('div', { 'class': 'vh-log-row', 'role': 'row' }, [
                E('time', { 'class': 'vh-log-time', 'role': 'cell' }, log.time || '时间未知'),
                E('span', { 'class': 'vh-level ' + log.level, 'role': 'cell' }, levelNames[log.level] || '信息'),
                E('span', { 'class': 'vh-log-message', 'title': log.detail || '', 'role': 'cell' }, log.message)
            ]);
        }) : [ E('div', { 'class': 'vh-empty' }, [ icon('logs'), E('span', {}, total ? '没有符合筛选条件的日志' : '暂无运行日志'), E('span', { 'class': 'vh-description' }, total ? '试试其他关键词或日志等级' : '服务产生新日志后会显示在这里') ]) ];
        this.logBody.replaceChildren.apply(this.logBody, nodes);
        this.logBody.scrollTop = scrollTop;
    },

    copyLogs: function() {
        var text = this.filteredLogs().map(function(log) { return log.time + ' [' + (levelNames[log.level] || '信息') + '] ' + log.message; }).join('\n');
        var copy;
        if (navigator.clipboard && window.isSecureContext) {
            copy = navigator.clipboard.writeText(text);
        } else {
            var input = E('textarea', { 'style': 'position:fixed;left:-9999px;top:0', 'readonly': true }, text);
            document.body.appendChild(input);
            input.select();
            var copied = false;
            try { copied = document.execCommand('copy'); } finally { input.remove(); }
            copy = copied ? Promise.resolve() : Promise.reject(new Error('复制失败'));
        }
        return copy.then(function() { ui.addNotification(null, E('p', {}, '已复制当前筛选的中文日志。'), 'info'); }).catch(function() {
            ui.addNotification(null, E('p', {}, '复制失败，请选择日志文字后手动复制。'), 'error');
        });
    },

    renderDiagnostics: function() {
        var info = this.snapshot.info;
        var line = function(label, value) { return E('div', { 'class': 'vh-diagnostic-line' }, [ E('span', {}, label), E('span', {}, value) ]); };
        var process = E('div', {}, [ E('h3', { 'class': 'vh-section-title' }, '进程与端口'), line('容器名称', info.container), line('进程编号', info.pid > 0 ? String(info.pid) : '暂无进程') ]);
        info.ports.forEach(function(port) { process.appendChild(line('监听端口 ' + port.port, port.address)); });
        if (!info.ports.length) process.appendChild(line('端口状态', '暂无监听端口'));
        var files = E('div', {}, [ E('h3', { 'class': 'vh-section-title' }, '服务文件') ]);
        info.files.forEach(function(file) {
            files.appendChild(E('div', { 'class': 'vh-file' }, [
                E('div', { 'class': 'vh-file-top' }, [ E('span', {}, file.label), E('span', { 'class': file.exists ? 'vh-file-ok' : 'vh-file-missing' }, file.exists ? '正常' : '不存在') ]),
                E('div', { 'class': 'vh-file-path' }, file.path)
            ]));
        });
        this.diagnostics.replaceChildren(process, files);
    },

    render: function(data) {
        this.readonly = !L.hasViewPermission();
        this.busy = false;
        this.pendingAction = null;
        this.snapshot = null;
        this.logSignature = null;
        this.actionButtons = {};
        this.cards = {};
        var self = this;
        var button = function(label, symbol, handler, extra) { return E('button', { 'type': 'button', 'class': 'vh-button ' + (extra || ''), 'click': handler }, [ icon(symbol), E('span', {}, label) ]); };
        var card = function(key, title, symbol, color) {
            var value = E('div', { 'class': 'vh-value' }, '—'), note = E('div', { 'class': 'vh-card-note' }, '正在读取状态');
            self.cards[key] = { value: value, note: note };
            if (key === 'status') {
                // This view owns disabled state; LuCI's handler disables before invoking us.
                self.powerButton = E('button', { 'type': 'button', 'class': 'vh-card-icon vh-card-button vh-power-toggle', 'disabled': true, 'aria-label': '等待服务状态', 'click': L.bind(self.toggleService, self) }, icon('power'));
                self.actionButtons.restart = E('button', { 'type': 'button', 'class': 'vh-card-icon vh-card-button vh-restart', 'disabled': true, 'title': '重启服务', 'aria-label': '重启服务', 'click': function() { return self.runAction('restart'); } }, icon('refresh'));
                return E('div', { 'class': 'vh-card vh-status-card' }, [
                    E('div', { 'class': 'vh-status-copy' }, [ E('div', { 'class': 'vh-card-top' }, title), value, note ]),
                    E('div', { 'class': 'vh-status-actions', 'role': 'group', 'aria-label': '服务启停与重启' }, [ self.powerButton, self.actionButtons.restart ])
                ]);
            }
            var cardIcon = E('span', { 'class': 'vh-card-icon ' + color }, icon(symbol));
            return E('div', { 'class': 'vh-card' }, [ E('div', { 'class': 'vh-card-top' }, [ E('span', {}, title), cardIcon ]), value, note ]);
        };
        this.notice = E('div', { 'class': 'vh-notice', 'role': 'status', 'hidden': true });
        this.badgeLabel = E('span', {}, '正在读取');
        this.badge = E('span', { 'class': 'vh-badge', 'role': 'status' }, [ E('span', { 'class': 'vh-dot' }), this.badgeLabel ]);
        this.refreshButton = E('button', { 'type': 'button', 'class': 'vh-card-icon vh-card-button', 'title': '刷新状态', 'aria-label': '刷新状态', 'click': ui.createHandlerFn(this, 'refreshStatus') }, icon('refresh'));
        this.searchInput = E('input', { 'type': 'search', 'placeholder': '搜索日志内容、设备或接口…', 'aria-label': '搜索日志', 'input': L.bind(this.renderLogs, this) });
        this.levelSelect = E('select', { 'class': 'vh-select', 'aria-label': '日志等级', 'change': L.bind(this.renderLogs, this) }, [ E('option', { 'value': '' }, '全部等级'), E('option', { 'value': 'error' }, '错误'), E('option', { 'value': 'warn' }, '警告'), E('option', { 'value': 'info' }, '信息'), E('option', { 'value': 'debug' }, '调试') ]);
        this.hideHeartbeat = E('input', { 'type': 'checkbox', 'checked': true, 'change': L.bind(this.renderLogs, this) });
        this.autoRefresh = E('input', { 'type': 'checkbox', 'checked': true });
        this.copyButton = button('复制日志', 'copy', ui.createHandlerFn(this, 'copyLogs'), 'vh-small');
        this.logBody = E('div', { 'class': 'vh-log-body', 'role': 'rowgroup', 'tabindex': '0', 'aria-label': '运行日志，可滚动查看' });
        this.logCount = E('span', {}, '正在读取日志');
        this.updated = E('span', {}, '等待更新');
        this.logNote = E('p', { 'class': 'vh-description' }, '最近 200 条记录 · 最新记录在前');
        this.diagnostics = E('div', { 'class': 'vh-diagnostic-content' });
        var page = E('div', { 'class': 'vh-page' }, [
            E('link', { 'rel': 'stylesheet', 'href': L.resource('view/services/vohive.css') + '?v=20261005-4' }),
            E('div', { 'class': 'vh-hero' }, [
                E('div', { 'class': 'vh-brand' }, [ E('div', { 'class': 'vh-logo' }, icon('hive')), E('div', {}, [ E('div', { 'class': 'vh-title' }, [ E('h2', { 'class': 'vh-heading' }, 'VoHive'), E('span', { 'class': 'vh-plus' }, 'PLUS') ]), E('p', { 'class': 'vh-subtitle' }, '设备互联，轻松掌控 · iStoreOS 服务管理') ]) ]),
                E('div', { 'class': 'vh-hero-tools' }, [ this.badge, E('a', { 'class': 'vh-button vh-primary', 'href': localManagementUrl(window.location.hostname), 'target': '_blank', 'rel': 'noopener noreferrer' }, [ E('span', {}, '本地管理'), icon('external') ]), E('a', { 'class': 'vh-button', 'href': 'https://xjp.721609.xyz/#/phone', 'target': '_blank', 'rel': 'noopener noreferrer' }, [ E('span', {}, '通话中心'), icon('external') ]) ])
            ]),
            this.notice,
            E('div', { 'class': 'vh-cards' }, [ card('status', '服务状态', 'power', 'green'), card('boot', '开机自启', 'check', ''), card('port', '访问端口', 'port', 'blue'), card('uptime', '运行时长', 'clock', 'amber') ]),
            E('div', { 'class': 'vh-panel' }, [
                E('div', { 'class': 'vh-log-header' }, [ E('div', {}, [ E('div', { 'class': 'vh-log-title' }, [ icon('logs'), E('h3', { 'class': 'vh-section-title' }, '运行日志') ]), this.logNote ]), E('div', { 'class': 'vh-log-tools' }, [ E('label', { 'class': 'vh-auto' }, [ this.autoRefresh, E('span', {}, '自动刷新 · 10 秒') ]), this.refreshButton ]) ]),
                E('div', { 'class': 'vh-toolbar' }, [ E('div', { 'class': 'vh-search' }, [ icon('search'), this.searchInput ]), this.levelSelect, E('label', { 'class': 'vh-auto' }, [ this.hideHeartbeat, E('span', {}, '隐藏心跳') ]), this.copyButton ]),
                E('div', { 'class': 'vh-log-table', 'role': 'table', 'aria-label': '中文运行日志' }, [ E('div', { 'class': 'vh-log-columns', 'role': 'row' }, [ E('span', { 'role': 'columnheader' }, '时间'), E('span', { 'role': 'columnheader' }, '等级'), E('span', { 'role': 'columnheader' }, '日志内容') ]), this.logBody ]),
                E('div', { 'class': 'vh-log-footer' }, [ this.logCount, this.updated ])
            ]),
            E('details', { 'class': 'vh-panel vh-diagnostics' }, [ E('summary', {}, [ E('span', { 'class': 'vh-summary-label' }, [ icon('shield'), E('span', {}, '服务诊断') ]), E('span', { 'class': 'vh-summary-hint' }, '查看进程、端口和文件'), icon('chevron') ]), this.diagnostics ]),
            E('div', { 'class': 'vh-bottom' }, [ E('span', { 'class': 'vh-dot' }), E('span', {}, 'VoHive Plus · 让连接更简单') ])
        ]);
        var background = window.getComputedStyle(document.body).backgroundColor.match(/[\d.]+/g);
        if (background && (background.length < 4 || +background[3] > 0) && (+background[0] * .299 + +background[1] * .587 + +background[2] * .114) < 110) page.setAttribute('data-theme', 'dark');
        if (data) this.updateStatus(data);
        else {
            this.notice.hidden = false;
            this.notice.textContent = '暂时无法读取服务状态，请检查连接或登录状态后重试。';
            this.badgeLabel.textContent = '状态未知';
            Object.keys(this.cards).forEach(function(key) { self.cards[key].note.textContent = '暂时无法读取'; });
            this.renderLogs();
            this.updateButtons();
        }
        if (this.pollCallback) poll.remove(this.pollCallback);
        this.pollCallback = L.bind(function() { if (this.autoRefresh.checked && !this.busy) return this.refreshStatus(); }, this);
        poll.add(this.pollCallback, 10);
        return page;
    },

    handleSaveApply: null,
    handleSave: null,
    handleReset: null
});
