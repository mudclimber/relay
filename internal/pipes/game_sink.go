package pipes


import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/mudclimber/relay/internal/pipes/terminalops"
	"github.com/mudclimber/relay/pkg/handler"
	"github.com/mudclimber/relay/pkg/sockets"
	"github.com/mudclimber/relay/pkg/agents"
)

type GameSink struct {
  Conn net.Conn
  FromSocket chan sockets.FromSocket
  ToSocket chan sockets.ToSocket
  ToAgent chan agents.ToAgent
  Initialize <-chan terminalops.TerminalOp
  HandlerInitDone chan []byte
  Login string
  EndItAll chan<- bool
}

func (s *GameSink) ReadChan(wg *sync.WaitGroup, handler handler.Handler) {
  Listen:
    for {
      msg := <-s.ToSocket
      switch msg.(type) {
      case []byte:
        log.Println(string([]byte(msg.([]byte))))
        log.Fatalln("Unexpected raw bytes:", msg)
      case sockets.SocketInput:
        _, err := s.Conn.Write(msg.(sockets.SocketInput).Buf)
        if err != nil {
          log.Println(err)
          break Listen
        }
      case sockets.SocketEOF:
        log.Println("TODO do something here")
      }
    }
  (*wg).Done()
}

func (s *GameSink) ReadForInitialize(wg *sync.WaitGroup, handler handler.Handler) {
  msg := (<-s.Initialize).(terminalops.Initialize)
  s.Login = msg.Login
  s.InitGame(handler)
  (*wg).Wait()
}

func (s *GameSink) ReadForGameOutput(wg *sync.WaitGroup, handler handler.Handler) {
  for {
    buf := make([]byte, 4096)
    n, err := s.Conn.Read(buf)
    if n == 0 {
      s.EndItAll <- true
      break
    }
    if err != nil {
      log.Println(err)
      break
    }
    log.Println("[game]", len(buf), n)
    log.Println("[game output sent to agent]")
    handler.ParseOutput(&buf)
    s.FromSocket <- sockets.SocketOutput {
      Buf: buf[:n],
    }
  }
  (*wg).Done()
}

func (s *GameSink) ReadUntilSize(n int, timeout int) bytes.Buffer {
  buf := bytes.Buffer{}
  count := 0
  for count < n {
    msg := make([]byte, 4096)
    _, err := s.Conn.Read(msg)
    if err != nil {
      log.Fatal(err)
    }
    index := bytes.Index(msg, []byte{0})
    if index < 0 {
      index = 4096
    }
    count += index
    buf.Write(msg[:index])
  }
  log.Printf("Stopped at %d bytes, got %d!\n", n, count)
  return buf
}

func (s *GameSink) Write(msg string) {
  if _, err := fmt.Fprintf(s.Conn, msg); err != nil {
    log.Println(err)
  }
}

func (s *GameSink) InitGame(h handler.Handler) error {
  c := handler.NewHandlerActions(s.FromSocket, s.ToSocket, s.ToAgent)
  defer func() {
    s.HandlerInitDone <- c.Intro
  }()
  return h.HandleInit(&c, s.Login)
}

func (s *GameSink) Run(h handler.Handler) {
  var gameWG sync.WaitGroup
  gameWG.Add(2)
  go s.ReadForInitialize(&gameWG, h)
  go s.ReadChan(&gameWG, h)

  gameWG.Add(1)
  go s.ReadForGameOutput(&gameWG, h)

  gameWG.Wait()
}
