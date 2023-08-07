package terminalops

import (
	pbmap "github.com/mudclimber/relay/internal/pipes/pbmap"
)

type TerminalOp interface{}

type Initialize struct {
  Login string
  Size pbmap.Geometry
}

type NewGeometry struct {
  Size pbmap.Geometry
}

type RedrawBox struct { }

type Clear struct { }

type AppendToPrompt struct {
  Buf []byte
}

type DeleteFromPrompt struct {
  NumBytes int
}

type ReplacePrompt struct {
  NewPrompt []byte
}

type AgentInput struct {
  Buf []byte
}

type ProcessPrompt struct { }
