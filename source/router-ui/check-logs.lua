local logs = dofile("ui/logs.lua")
local function check(name, line, expected)
    local value = logs.parse(line)
    for key, wanted in pairs(expected) do
        assert(value[key] == wanted, name .. ": " .. key .. " = " .. tostring(value[key]))
    end
    assert(not value.message:find("\27", 1, true), "ANSI escape remains")
    print("通过：" .. name)
end

check("中文请求摘要", 'Thu Oct  1 16:39:22 2026 daemon.info vohive-plus-run[1910]: [GIN] 2026/10/01 - 16:39:22 | 200 | 238.331µs | 192.168.100.106 | GET "/api/devices"', {
    time = "2026-10-01 16:39:22", level = "info", heartbeat = false,
    message = "读取设备列表 · 状态 200 · 耗时 238.331 微秒 · 来源 192.168.100.106"
})
check("短信接口和查询参数", '[GIN] 2026/10/01 - 16:39:22 | 200 | 1.28ms | 192.168.100.106 | GET "/api/sms/notifications?after_id=103"', {
    detail = "/api/sms/notifications?after_id=103", heartbeat = false,
    message = "读取短信通知 · 状态 200 · 耗时 1.28 毫秒 · 来源 192.168.100.106"
})
check("服务心跳", '[GIN] 2026/10/01 - 16:39:22 | 200 | 61µs | 127.0.0.1 | GET "/ping"', { heartbeat = true })
check("心跳筛选不误伤其他接口", '[GIN] 2026/10/01 - 16:39:22 | 200 | 61µs | 127.0.0.1 | GET "/pingpong"', { heartbeat = false })
check("服务错误等级", '[GIN] 2026/10/01 - 16:39:22 | 500 | 1ms | ::1 | POST "/api/devices"', { level = "error" })
check("请求警告等级", '[GIN] 2026/10/01 - 16:39:22 | 404 | 1ms | ::1 | GET "/missing"', { level = "warn" })
check("去掉控制字符并翻译字段", 'Thu Oct  1 16:39:24 2026 daemon.info vohive-plus-run[1910]: [2026-10-01 16:39:24] \27[34mINFO \27[0m device/pool_health_ip.go:261 [黑色ec20] 缺少 Worker {"imei": "123", "active_workers": 2}', {
    time = "2026-10-01 16:39:24", level = "info", detail = "device/pool_health_ip.go:261",
    message = '[黑色ec20] 缺少 工作进程 {"设备标识": "123", "工作进程数": 2}'
})
check("系统错误和中文时间", 'Thu Oct  1 16:39:24 2026 daemon.err vohive-plus-run[1910]: connection refused', {
    time = "2026-10-01 16:39:24", level = "error", message = "连接被拒绝"
})
check("未识别消息原样保留", '[2026-10-01 16:39:24] WARN service/main.go:10 未识别的设备状态 <script>alert(1)</script>', {
    level = "warn", message = "未识别的设备状态 <script>alert(1)</script>"
})
check("异常健康状态", 'Thu Oct  1 16:39:24 2026 daemon.info vohive-plus-run[1910]: unhealthy', { message = "异常" })
