package sandbox

import (
	"goraven/core/sandbox/types"
)

type (
	Sandbox         = types.Sandbox
	StorageUsage    = types.StorageUsage
	StorageCapacity = types.StorageCapacity
	FileManager     = types.FileManager
)

func NewSandbox(userName string) (Sandbox, error) {
	return nil, nil

}
