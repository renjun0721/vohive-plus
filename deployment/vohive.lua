module("luci.controller.vohive", package.seeall)

function index()
    entry({"admin", "services", "vohive"}, view("services/vohive"), _("VoHive Plus"), 60).acl_depends = { "luci-app-vohive" }
end
