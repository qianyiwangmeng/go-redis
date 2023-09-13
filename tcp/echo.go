package tcp

import "sync"

type EchoHandler struct {
	activeConn sync.Map
}
