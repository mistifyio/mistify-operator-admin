package models

import "net"

type Hypervisor struct {
	ID       string            `json:"id"`
	Mac      net.HardwareAddr  `json:"mac"`
	IPv6     net.IP            `json:"ipv6"`
	Metadata map[string]string `json:"metadata"`
	IPRanges []*IPRange        `json:"-"`
}
