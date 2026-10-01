//go:build !linux

package personal

func xfrmCheck() Check { return Check{"xfrm_ipsec", false, "Requires a Linux host"} }
