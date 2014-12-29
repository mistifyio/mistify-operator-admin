package models

import "errors"

var ErrNoID = errors.New("missing id")
var ErrBadID = errors.New("invalid id")
var ErrNilMetadata = errors.New("metadata must not be nil")
var ErrNoName = errors.New("missing name")
var ErrNoMAC = errors.New("missing MAC")
var ErrNoIP = errors.New("missing IP")
var ErrNoService = errors.New("missing service")
var ErrNoAction = errors.New("missing action")
var ErrNoUsername = errors.New("missing username")
var ErrNoEmail = errors.New("missing email")
var ErrNoCIDR = errors.New("missing cidr")
var ErrNoGateway = errors.New("missing gateway")
var ErrNoStartIP = errors.New("missing start IP")
var ErrNoEndIP = errors.New("missing end IP")
var ErrBadCPU = errors.New("cpu must be > 0")
var ErrBadMemory = errors.New("memory must be > 0")
var ErrBadDisk = errors.New("disk must be > 0")
var ErrNilData = errors.New("data must not be nil")
