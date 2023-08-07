package rpc

import (
	"io"
	"log"
	"net"

	"github.com/mudclimber/relay/internal/pipes"
	"github.com/mudclimber/relay/internal/pipes/pbmap"
	"github.com/mudclimber/relay/internal/pipes/terminalops"
	pb "github.com/mudclimber/relay/internal/proto/hub_agent_protos"
	"github.com/mudclimber/relay/pkg/agents"
	"github.com/mudclimber/relay/pkg/handler"
	"github.com/mudclimber/relay/pkg/sockets"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)


type Geometry struct {
  rows int
  columns int
}

func (g Geometry) plus(rows int, columns int) Geometry {
  return Geometry {
    rows: g.rows + rows,
    columns: g.columns + columns,
  }
}

type EvenniaState int

// gonna change. just want something here for now to make the room not feel empty
const (
  Auth EvenniaState = iota
  Playing
)

type imember interface {
  initEvennia() error
  Read(buf []byte)
  Write(buf []byte)
}

func loginGeometry(frame *pb.Login) Geometry {
  return Geometry {
    rows: int(frame.Rows),
    columns: int(frame.Columns),
  }
}

/* NOTE

I want to decouple this with channels:

In:
  - login frames, read frames, resize frames, disconnect frames

On:
  - resize, send bytes to game

Out:
  - opChan: operation (remove bytes, append bytes, process buffer)
  - bytes to game

The goroutine:
  - on process buffer recv, send buffer to evennia (another channel?)
*/

type rpcInput struct {
  serverIn chan<- pbmap.ProtoMap
  stream *pb.AgentStreams_AgentStreamServer
}

func (r *rpcInput) Run() {
  for {
    request, err := (*r.stream).Recv()
    if err != nil {
      if status.Code(err) != codes.Canceled {
        log.Println(err)
      } else if err == io.EOF {
        log.Println("Game connection ended on me (EOF)")
      } else {
        log.Println("RPC Connection ended")
      }
      break
    }
    frameMsg, err := pbmap.Convert(request)
    if err != nil {
      log.Println(err)
      break
    }
    r.serverIn <- frameMsg
  }
}

type Servicer struct {
  pb.UnimplementedAgentStreamsServer

  Handler handler.Handler
}

func (s *Servicer) AgentStream(stream pb.AgentStreams_AgentStreamServer) error {
  log.SetFlags(log.LstdFlags | log.Lshortfile)
  conn, err := net.Dial("tcp", "localhost:4000")
  if err != nil {
    log.Panicln("No connection to port 4000")
  }
  defer conn.Close()

  serverOut := make(chan agents.ToAgent)
  terminal := make(chan terminalops.TerminalOp)
  serverIn := make(chan pbmap.ProtoMap)
  toSocket := make(chan sockets.ToSocket)
  fromSocket := make(chan sockets.FromSocket)
  socketInit := make(chan terminalops.TerminalOp)
  endItAll := make(chan bool)
  handlerInitDone := make(chan []byte)

  frameProcessor := pipes.FrameProcessor {
    Frame: serverIn,
    Terminal: terminal,
  }

  terminalProcessor := pipes.TerminalProcessor {
    Terminal: terminal,
    ToAgent: serverOut,
    ToSocket: toSocket,
    FromSocket: fromSocket,
    GameInitialize: socketInit, // to send the init message to game sink
    HandlerInitDone: handlerInitDone,
  }

  agentSink := pipes.AgentSink {
    Stream: &stream,
    Response: serverOut,
  }

  gameSink := pipes.GameSink {
    Conn: conn,
    FromSocket: fromSocket,
    ToSocket: toSocket,
    ToAgent: serverOut,
    Initialize: socketInit, // relays the init to the game to process init/login info
    HandlerInitDone: handlerInitDone,
    EndItAll: endItAll,
  }

  rpcIn := rpcInput {
    serverIn: serverIn,
    stream: &stream,
  }

  go gameSink.Run(s.Handler) // Handler needed for init interception
  go agentSink.Run(s.Handler) // Handler needed for parsing output
  go frameProcessor.Run()
  go terminalProcessor.Run()
  go rpcIn.Run()

  <-endItAll

  return nil
}

