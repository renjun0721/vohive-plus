-- Chinese presentation of service logs. Stored application logs remain intact.
local M = {}

local levels = { INFO = "info", NOTICE = "info", WARN = "warn", WARNING = "warn", ERR = "error", CRIT = "error", ALERT = "error", EMERG = "error", ERROR = "error", FATAL = "error", PANIC = "error", DEBUG = "debug", TRACE = "debug" }
local methods = { GET = "读取", POST = "提交", PUT = "更新", PATCH = "修改", DELETE = "删除", OPTIONS = "预检", HEAD = "检查" }
local routes = {
    { "^/api/sms/notifications", "短信通知" },
    { "^/api/dashboard/devices", "设备概览" },
    { "^/api/dashboard", "控制台概览" },
    { "^/api/devices", "设备列表" },
    { "^/api/sms", "短信" },
    { "^/api/settings", "设置" },
    { "^/api/config", "配置" },
    { "^/api/contacts", "联系人" },
    { "^/api/calls", "通话记录" },
    { "^/api/login", "登录接口" },
    { "^/ping", "服务心跳" },
    { "^/$", "管理页面" }
}
local words = {
    { "active_workers", "工作进程数" }, { "activeWorkers", "工作进程数" },
    { "Worker", "工作进程" }, { "worker", "工作进程" },
    { "imei", "设备标识" }, { "signal", "信号" },
    { "terminated", "已终止" }, { "context canceled", "操作已取消" },
    { "connection refused", "连接被拒绝" }, { "connection reset by peer", "对端重置连接" },
    { "no such file or directory", "文件或目录不存在" },
    { "permission denied", "权限不足" }, { "i/o timeout", "读写超时" },
    { "deadline exceeded", "操作超时" }, { "address already in use", "端口已被占用" },
    { "Starting", "正在启动" }, { "Stopping", "正在停止" },
    { "Listening", "正在监听" }, { "listening", "正在监听" },
    { "Server", "服务" }, { "server", "服务" },
    { "failed", "失败" }, { "Failed", "失败" },
    { "unhealthy", "异常" }, { "healthy", "正常" }
}

local function trim(s)
    return (s:gsub("^%s+", ""):gsub("%s+$", ""))
end

local function chinese(s)
    for _, pair in ipairs(words) do
        s = s:gsub(pair[1], pair[2])
    end
    return s
end

local function chinese_duration(s)
    return trim(s):gsub("µs", " 微秒"):gsub("μs", " 微秒"):gsub("ns", " 纳秒"):gsub("ms", " 毫秒")
        :gsub("([%d%.]+)h", "%1 小时 "):gsub("([%d%.]+)m", "%1 分 "):gsub("([%d%.]+)s", "%1 秒")
end

function M.parse(line)
    line = line:gsub("\27%[[%d;]*[A-Za-z]", ""):gsub("[%z\1-\8\11\12\14-\31]", "")
    local body = line:match("vohive[^:]*:%s*(.*)") or line
    local timestamp = body:match("(%d%d%d%d[/-]%d%d[/-]%d%d%s+%d%d:%d%d:%d%d)")
        or body:match("(%d%d%d%d/%d%d/%d%d%s*%-%s*%d%d:%d%d:%d%d)")
    if timestamp then
        timestamp = timestamp:gsub("/", "-"):gsub("%s+%-%s+", " ")
    else
        local month, day, clock, year = line:match("^%a+%s+(%a+)%s+(%d+)%s+(%d%d:%d%d:%d%d)%s+(%d%d%d%d)")
        local months = { Jan=1, Feb=2, Mar=3, Apr=4, May=5, Jun=6, Jul=7, Aug=8, Sep=9, Oct=10, Nov=11, Dec=12 }
        timestamp = months[month] and string.format("%s-%02d-%02d %s", year, months[month], tonumber(day), clock) or ""
    end

    local status, duration, client, method, path = body:match("%[GIN%].-|%s*(%d+)%s*|%s*(.-)%s*|%s*(.-)%s*|%s*(%u+)%s+\"([^\"]+)\"")
    if status then
        local code, target = tonumber(status), "接口请求"
        for _, route in ipairs(routes) do
            if path:match(route[1]) then target = route[2]; break end
        end
        return {
            time = timestamp, level = code >= 500 and "error" or code >= 400 and "warn" or "info",
            message = string.format("%s%s · 状态 %s · 耗时 %s · 来源 %s", methods[method] or "访问", target, status, chinese_duration(duration), trim(client)),
            detail = path, heartbeat = path == "/ping" or path:match("^/ping[?/]") ~= nil
        }
    end

    local level, source, message = body:match("^%[.-%]%s+(%u+)%s+([^%s]+%.go:%d+)%s+(.*)")
    if not level then
        level, message = body:match("^%[.-%]%s+(%u+)%s+(.*)")
    end
    if not message then
        message = body
        level = line:match("daemon%.(%a+)") or "info"
        level = level:upper()
    end
    return { time = timestamp, level = levels[level] or "info", message = chinese(trim(message)), detail = source or "", heartbeat = false }
end

return M
