package handler

import (
	"github.com/mudclimber/relay/pkg/agents"
	"github.com/mudclimber/relay/pkg/sockets"
  "bytes"
  "log"
)

/*
Todo:
- eventually write some godoc, and some static markdown
  to host on the README.
*/

type Handler interface {
	HandleInit(*HandlerActions, string) error
	ParseOutput(*[]byte)
}

//////////////////////////////////////////

type HandlerOptions struct {
	DisplayName string
	Port        uint32
}

//////////////////////////////////////////

type HandlerIn interface{}

type HandlerRead struct {
	Buf []byte
}

type HandlerOut interface{}

type HandlerWrite struct {
	Buf []byte
}


//////////////////////////////////////////

type HandlerActions struct {
  fromSocket <-chan sockets.FromSocket
  toSocket chan<- sockets.ToSocket
  Intro []byte
}

func NewHandlerActions(
  fromSocket <-chan sockets.FromSocket,
  toSocket chan<- sockets.ToSocket,
  toAgent chan agents.ToAgent,
) HandlerActions {
  return HandlerActions {
    fromSocket: fromSocket,
    toSocket: toSocket,
  }
}

func (a *HandlerActions) ReadUntilSize(n int, timeout int) bytes.Buffer {
  buf := bytes.Buffer{}
  count := 0
  for count < n {
    log.Println("you trying to read something?")
    msg := <-a.fromSocket
    switch msg.(type) {
    case sockets.SocketOutput:
      smsg := msg.(sockets.SocketOutput)
      log.Println("GOT IT")
      index := bytes.Index(smsg.Buf, []byte{0})
      if index < 0 {
        index = len(smsg.Buf)
      }
      count += index
      buf.Write(smsg.Buf[:index])
      log.Printf("Stopped at %d bytes, got %d!\n", n, count)
    default:
      log.Fatalln("NO", msg)
    }
  }
  return buf
}

func (a *HandlerActions) SendBytes(buf []byte) {
  a.toSocket <- sockets.SocketInput {
    Buf: buf,
  }
}
