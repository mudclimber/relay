package agents

// Interface for messages pass through channels
type FromAgent interface{}

// for FromAgent
type AgentRead struct {
  Buf []byte
}

// for FromAgent
type AgentClose struct{}

//////////////////////////////////////////

type ToAgent interface{}

// for ToAgent
type AgentWrite struct {
  Buf []byte
}

// for ToAgent
type AgentEOF struct{}
