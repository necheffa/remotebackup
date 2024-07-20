package data

import (
	"errors"
)

var (
	ErrUnsupportedVolumeType    = errUnsupportedVolumeType()
	ErrUnsupportedFileSystem    = errUnsupportedFileSystem()
	ErrUnableToCreateMountPoint = errUnableToCreateMountPoint()
)

type Snapshot interface {
	Create() error
	Destroy() error
	Mount() error
	Unmount() error
	Volume() *Volume
	Host() *Host
}

func NewSnapshot(conf *Configuration, conn Connection, h *Host, v *Volume) (Snapshot, error) {
	if v.Type == "lvm" {
		return NewLvmSnapshot(conf, conn, h, v)
	}

	return nil, ErrUnsupportedVolumeType
}

func errUnsupportedVolumeType() error {
	return errors.New("unsupported volume type")
}

func errUnsupportedFileSystem() error {
	return errors.New("unsupported filesystem")
}

func errUnableToCreateMountPoint() error {
	return errors.New("unable to create mountpoint")
}
