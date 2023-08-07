package pipes

import (
	pbmap "github.com/mudclimber/relay/internal/pipes/pbmap"
	"github.com/mudclimber/relay/internal/pipes/terminalops"
)

// Agent frame (mapped) from protobuf converted to terminal operations

type FrameProcessor struct {
  Frame <-chan pbmap.ProtoMap
  Terminal chan<- terminalops.TerminalOp
  promptBuffer []byte // TODO: does this make sense here? I guess so
}

func (f *FrameProcessor) Run() {
  for {
    frame := <-f.Frame
    switch frame.(type) {
    case pbmap.Login:
      loginFrame := frame.(pbmap.Login)
      f.Terminal <- terminalops.Initialize {
        Login: loginFrame.Login,
        Size: loginFrame.Geometry,
      }
    case pbmap.Read:
      readFrame := frame.(pbmap.Read)
      f.Terminal <- terminalops.AgentInput {
        Buf: readFrame.Buf,
      }
    case pbmap.Resize:
      resizeFrame := frame.(pbmap.Resize)
      f.Terminal <- terminalops.NewGeometry{
        Size: pbmap.Geometry {
          Rows: resizeFrame.NewSize.Rows,
          Columns: resizeFrame.NewSize.Columns,
        },
      }
      f.Terminal <- terminalops.RedrawBox { }
      f.Terminal <- terminalops.ReplacePrompt {
        NewPrompt: f.promptBuffer,
      }
    }
  }
}
