'use strict';
'require view';
'require rpc';
'require fs';
'require ui';
'require poll';

var callRcList = rpc.declare({
    object: 'rc',
    method: 'list',
    params: [ 'name' ],
    expect: { '': {} }
});

var callRcInit = rpc.declare({
    object: 'rc',
    method: 'init',
    params: [ 'name', 'action' ],
    expect: { result: false }
});

return view.extend({
    load: function() {
        return Promise.all([
            L.resolveDefault(callRcList('vohive'), {}),
            L.resolveDefault(fs.exec('/usr/libexec/vohive-luci-info'), { stdout: '', stderr: '' })
        ]);
    },

    runAction: function(action) {
        return callRcInit('vohive', action).then(L.bind(function() {
            ui.addNotification(null, E('p', _('VoHive command sent: %s').format(action)), 'info');
            return this.refreshStatus();
        }, this)).catch(function(e) {
            ui.addNotification(null, E('p', e.message || _('Command failed')));
        });
    },

    refreshStatus: function() {
        return Promise.all([
            L.resolveDefault(callRcList('vohive'), {}),
            L.resolveDefault(fs.exec('/usr/libexec/vohive-luci-info'), { stdout: '', stderr: '' })
        ]).then(L.bind(function(data) {
            this.updateStatus(data[0], data[1]);
        }, this));
    },

    updateStatus: function(service, info) {
        var svc = service && service.vohive ? service.vohive : {};
        var inst = svc.instances || {};
        var running = !!svc.running;

        Object.keys(inst).forEach(function(k) {
            if (inst[k].running)
                running = true;
        });

        var badge = document.getElementById('vohive-status-badge');
        var details = document.getElementById('vohive-status-details');

        if (badge) {
            badge.textContent = running ? _('Running') : _('Stopped');
            badge.className = running ? 'ifacebadge large' : 'ifacebadge large interface-disabled';
        }

        if (details)
            details.textContent = (info && (info.stdout || info.stderr)) ? ((info.stdout || '') + (info.stderr || '')) : _('No status output available.');
    },

    render: function(data) {
        var service = data[0], info = data[1];
        var viewNode = E('div', { 'class': 'cbi-map' }, [
            E('h2', _('VoHive')),
            E('div', { 'class': 'cbi-map-descr' }, _('Manage the VoHive service running directly on iStoreOS.')),
            E('div', { 'class': 'cbi-section' }, [
                E('div', { 'style': 'display:flex;align-items:center;gap:12px;flex-wrap:wrap;margin-bottom:16px' }, [
                    E('span', { 'id': 'vohive-status-badge', 'class': 'ifacebadge large interface-disabled' }, _('Loading')),
                    E('button', { 'class': 'btn cbi-button cbi-button-apply', 'click': ui.createHandlerFn(this, 'runAction', 'start') }, _('Start')),
                    E('button', { 'class': 'btn cbi-button cbi-button-reload', 'click': ui.createHandlerFn(this, 'runAction', 'restart') }, _('Restart')),
                    E('button', { 'class': 'btn cbi-button cbi-button-reset', 'click': ui.createHandlerFn(this, 'runAction', 'stop') }, _('Stop')),
                    E('button', { 'class': 'btn cbi-button cbi-button-action', 'click': ui.createHandlerFn(this, 'refreshStatus') }, _('Refresh')),
                    E('a', { 'class': 'btn cbi-button cbi-button-link', 'href': 'http://' + window.location.hostname + ':7575/', 'target': '_blank', 'rel': 'noopener' }, _('Open Web UI'))
                ]),
                E('pre', { 'id': 'vohive-status-details', 'style': 'min-height:22em;white-space:pre-wrap;word-break:break-word' }, '')
            ])
        ]);

        setTimeout(L.bind(function() { this.updateStatus(service, info); }, this), 0);
        poll.add(L.bind(this.refreshStatus, this), 10);

        return viewNode;
    },

    handleSaveApply: null,
    handleSave: null,
    handleReset: null
});
