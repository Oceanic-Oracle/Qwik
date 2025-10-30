package server

type Server interface {
	CreateServer()
	Close() error
}
