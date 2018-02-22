package api

const (
	// DeviceEnabled represents enabled
	DeviceEnabled = "Enabled"

	// DeviceDisabled represents disabled
	DeviceDisabled = "Disabled"

	// DeviceFailed represents failed
	DeviceFailed = "Failed"
)

// AddDeviceReq structure
type AddDeviceReq struct {
	Devices []string `json:"devices"`
}

// DeviceAddResp is the success response sent to a AddDeviceReq request
type DeviceAddResp Peer
