package sockets

type FromSocket interface{}

// for FromSocket
type SocketOutput struct {
  Buf []byte
}

// for FromSocket
type SocketClose struct{}

//////////////////////////////////////////

type ToSocket interface{}

// for ToSocket
type SocketInput struct {
  Buf []byte
}

// for ToSocket
type SocketEOF struct{}

