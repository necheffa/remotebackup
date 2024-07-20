package data

type Connection interface {
	Bind(mntCmd, remoteMount, localMount string) error
	Unbind(remoteMount, localMount string) error
}
