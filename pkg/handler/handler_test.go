package handler

import (
  "testing"
	agents "github.com/mudclimber/relay/pkg/agents"
)

func TestAgentStructs(t *testing.T) {
  t.Log("hi")
  t.Log(agents.AgentRead {
    Buf: []byte{},
  })
  t.Log(agents.AgentWrite {
    Buf: []byte{},
  })
}
