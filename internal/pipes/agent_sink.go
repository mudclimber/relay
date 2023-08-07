package pipes

import (
	"log"

	pb "github.com/mudclimber/relay/internal/proto/hub_agent_protos"
	"github.com/mudclimber/relay/pkg/handler"
	"github.com/mudclimber/relay/pkg/agents"
)

type AgentSink struct {
  Stream *pb.AgentStreams_AgentStreamServer
  Response <-chan agents.ToAgent
}

func (s *AgentSink) Run(h handler.Handler) {
  for {
    resp := <-s.Response
    switch resp.(type) {
    case agents.AgentWrite:
      responseProto := pb.AgentStreamResponse {
        Write: &pb.Write {
          Buf: resp.(agents.AgentWrite).Buf,
        },
      }
      if err := (*s.Stream).Send(&responseProto); err != nil {
        log.Printf("[Agent] send failed: %s\n", err)
      }
    case agents.AgentClose:
    default:
      panic("Unexpected type for response chan.")
    }
  }
}
