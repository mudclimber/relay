package server

/*
This runs all the goroutines necessary to make
everything function. Main entrypoint is Run
*/

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"
  "sync"

	pb "github.com/mudclimber/relay/internal/proto/hub_agent_protos"
	"github.com/mudclimber/relay/internal/rpc"
	"github.com/mudclimber/relay/pkg/handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var HUB_ENDPOINT = "localhost:10101"

type agentData struct {
  client pb.AgentCommsClient
  displayName string
  port uint32
  handler handler.Handler
}

type MuxServer struct { }

func (m MuxServer) registerAgent(agentData agentData) []byte {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    req := pb.NewAgentRequest {
      Agent: &pb.Agent {
        Connection: &pb.Agent_Rpc {
          Rpc: &pb.RpcConnection {
            Host: "localhost",
            Port: agentData.port,
          },
        },
        DisplayName: agentData.displayName,
      },
    }
    stream, err := agentData.client.NewAgent(ctx, &req)
    if err != nil {
      panic(err)
    }
    return stream.GetGuid()
}

func (m MuxServer) pingIntervals(client pb.AgentCommsClient, guid []byte) {
  tick := time.Tick(5000 * time.Millisecond);
  for {
    <-tick;
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    req := pb.PingRequest{
      Disconnect: false,
      Guid: guid,
    }
    _, err := client.PingAgent(ctx, &req)
    if err != nil {
      panic(err)
    }
    // fmt.Println("Ping done")
  }
}

func (m *MuxServer) sessionsRpcConnection() *grpc.ClientConn {
	var opts []grpc.DialOption
  opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

  log.Println("Initializing RPC...")
  conn, err := grpc.Dial(HUB_ENDPOINT, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
  return conn
}

func (m *MuxServer) runPings(agentData agentData) {
  agentGuid := m.registerAgent(agentData)
  m.pingIntervals(agentData.client, agentGuid)
}

func (m *MuxServer) runStreams(agentData agentData) {
  grpcServer := grpc.NewServer()
  servicer := rpc.Servicer {
    Handler: agentData.handler,
  }
  pb.RegisterAgentStreamsServer(grpcServer, &servicer)

  lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", agentData.port))
  if err != nil {
    log.Fatalf("failed to listen: %v", err)
  }
  grpcServer.Serve(lis)
}

func (m *MuxServer) Run(handler handler.Handler, opts handler.HandlerOptions) {
  conn := m.sessionsRpcConnection()
	defer conn.Close()

  agentData := agentData {
    client: pb.NewAgentCommsClient(conn),
    displayName: opts.DisplayName,
    handler: handler,
    port: opts.Port,
  }

  var wg sync.WaitGroup
  wg.Add(2)
  go m.runPings(agentData)
  go m.runStreams(agentData)
  wg.Wait()
}
