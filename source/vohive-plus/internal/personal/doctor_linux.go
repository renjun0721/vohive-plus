//go:build linux

package personal

import "golang.org/x/sys/unix"

func xfrmCheck() Check {
	socket, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW, unix.NETLINK_XFRM)
	if err != nil {
		return Check{"xfrm_ipsec", false, err.Error()}
	}
	unix.Close(socket)
	return Check{"xfrm_ipsec", true, "NETLINK_XFRM available; crypto algorithms require separate verification"}
}
