package models

import "errors"

// ErrNoID is for a missing id
var ErrNoID = errors.New("missing id")

// ErrBadID is for an invalid id (e.g. non-uuid)
var ErrBadID = errors.New("invalid id")

// ErrNilMetadata is for nil metadata
var ErrNilMetadata = errors.New("metadata must not be nil")

// ErrNoName is for a missing name
var ErrNoName = errors.New("missing name")

// ErrNoMAC is for a missing MAC in the hypervisor
var ErrNoMAC = errors.New("missing MAC")

// ErrNoIP is for a missing IP in the hypervisor
var ErrNoIP = errors.New("missing IP")

// ErrNoService is for calling an unconfigured service
var ErrNoService = errors.New("missing service")

// ErrNoAction is for calling an unconfigured action
var ErrNoAction = errors.New("missing action")

// ErrNoUsername is for missing a username in the user object
var ErrNoUsername = errors.New("missing username")

// ErrNoEmail is for missing an email in the user object
var ErrNoEmail = errors.New("missing email")

// ErrNoCIDR is for missing a cidr in the iprange
var ErrNoCIDR = errors.New("missing cidr")

// ErrNoGateway is for missing a gateway in the iprange
var ErrNoGateway = errors.New("missing gateway")

// ErrNoStartIP is for missing a starting ip in the iprange
var ErrNoStartIP = errors.New("missing start IP")

// ErrNoStopIP is for missing a stopping ip in the iprange
var ErrNoEndIP = errors.New("missing end IP")

// ErrBadCPU is for an invlid CPU in the flavor
var ErrBadCPU = errors.New("cpu must be > 0")

// ErrBadMemory is for invalid memory in the flavor
var ErrBadMemory = errors.New("memory must be > 0")

// ErrBadDisk is for invalid disk in the flavor
var ErrBadDisk = errors.New("disk must be > 0")

// ErrNilData is for nil data map in the config
var ErrNilData = errors.New("data must not be nil")
