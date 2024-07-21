package data

import (
	"errors"
)

var (
	ErrBindFailure   = errBindFailure()
	ErrUnbindFailure = errUnbindFailure()
)

type Connection interface {
	Bind(mntCmd, remoteMount, localMount string) error
	Unbind(remoteMount, localMount string) error
}

func errBindFailure() error {
	return errors.New("bind")
}

func errUnbindFailure() error {
	return errors.New("unbind")
}
