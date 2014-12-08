package models

import "net"

type IPRange struct {
	ID       string            `json:"id"`
	CIDR     net.IPNet         `json:"cidr"`
	Gateway  net.IP            `json:"gateway"`
	Start    net.IP            `json:"start"`
	End      net.IP            `json:"end"`
	Metadata map[string]string `json:"metadata"`
	Network  *Network          `json:"-"`
}
