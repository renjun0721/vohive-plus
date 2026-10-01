package personal

import (
	"github.com/yibaiba/hideck/internal/audiotranscode"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var diagnosticCodecs = audiotranscode.New()

type Check struct {
	Key    string `json:"key"`
	Ready  bool   `json:"ready"`
	Detail string `json:"detail"`
}

// Doctor performs read-only host checks, inspired by VoCat's environment
// diagnostics. It does not install packages, bind USB drivers or send AT commands.
func Doctor() map[string]any {
	ports, _ := filepath.Glob("/dev/ttyUSB*")
	controls, _ := filepath.Glob("/dev/cdc-wdm*")
	qmiPath, qmiErr := exec.LookPath("qmi-proxy")
	if qmiErr != nil {
		if info, err := os.Stat("/usr/libexec/qmi-proxy"); err == nil && info.Mode()&0111 != 0 {
			qmiPath, qmiErr = "/usr/libexec/qmi-proxy", nil
		}
	}
	qmicli, qmicliErr := exec.LookPath("qmicli")
	manager := "manual"
	for _, command := range []string{"apk", "opkg", "apt-get"} {
		if _, err := exec.LookPath(command); err == nil {
			manager = command
			break
		}
	}
	checks := []Check{
		{"at_ports", len(ports) > 0, "ttyUSB port discovery"},
		{"qmi_controls", len(controls) > 0, "cdc-wdm control discovery"},
		{"qmi_proxy", qmiErr == nil, qmiPath},
		{"qmicli", qmicliErr == nil, qmicli},
		xfrmCheck(),
	}
	capabilities := diagnosticCodecs.Capabilities()
	for _, item := range []struct{ key, codec string }{{"codec_amr", "AMR"}, {"codec_amrwb", "AMR-WB"}, {"codec_mp3", "MP3"}} {
		detail := capabilities[item.codec]
		ready := detail == ""
		if ready {
			detail = "native codec symbols loaded"
		}
		checks = append(checks, Check{item.key, ready, detail})
	}
	return map[string]any{
		"product": "VoHive Plus Personal", "base": "HiDeck 2.1.23",
		"platform":        runtime.GOOS + "/" + runtime.GOARCH,
		"package_manager": manager, "serial_port_count": len(ports),
		"qmi_control_count": len(controls), "checks": checks,
		"read_only": true,
	}
}
