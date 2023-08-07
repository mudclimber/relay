package pipes

import (
	"testing"
  "time"

	"github.com/mudclimber/relay/internal/pipes/pbmap"
	"github.com/mudclimber/relay/internal/pipes/terminalops"
)

func TestFramePipe(t *testing.T) {
  afch := make(chan pbmap.ProtoMap)
  toch := make(chan terminalops.TerminalOp)
  ftt := FrameProcessor {
    Frame: afch,
    Terminal: toch,
  }
  go ftt.Run()
  go func(t *testing.T) {
    time.Sleep(5 * time.Second)
    panic("The tests shouldn't take this long")
  }(t)

  {
    afch <- pbmap.Login {
      Login: "mytest",
      Geometry: pbmap.Geometry {
        Rows: 24,
        Columns: 80,
      },
    }
    response := (<-toch).(terminalops.Initialize)
    if response.Size.Rows != 24 {
      t.Fatalf("Expected 24 rows")
    }
    if response.Size.Columns != 80 {
      t.Fatalf("Expected 80 columns")
    }
    if response.Login != "mytest" {
      t.Fatalf("Expected 'mytest' login")
    }
  }

  {
    afch <- pbmap.Read {
      Buf: []byte{'a', 127, 'b'},
    }

    // TODO run these in check timeouts
    // response := (<-toch).(terminalops.ReplacePrompt)
    // if len(response.NewPrompt) != 1 {
    //   t.Fatalf("buf len wrong size")
    // }
    // if response.NewPrompt[0] != 'a' {
    //   t.Fatalf("buf should be 'a' (%s)", string(response.NewPrompt))
    // }
    // response = (<-toch).(terminalops.ReplacePrompt)
    // if len(response.NewPrompt) != 0 {
    //   t.Fatalf("buf len wrong size")
    // }
    // response = (<-toch).(terminalops.ReplacePrompt)
    // if len(response.NewPrompt) != 1 {
    //   t.Fatalf("buf len wrong size (%s)", string(response.NewPrompt))
    // }
    // if response.NewPrompt[0] != 'b' {
    //   t.Fatalf("buf should be 'b' (%s)", string(response.NewPrompt))
    // }
    // afch <- pbmap.Read {
    //   Buf: []byte{127, 127},
    // }
    // response = (<-toch).(terminalops.ReplacePrompt)
    // if len(response.NewPrompt) != 0 {
    //   t.Fatalf("buf len wrong size (%s)", string(response.NewPrompt))
    // }
    // if len(toch) > 0 {
    //   t.Fatalf("Delete on empty prompt buffer shouldn't emit a message")
    // }
  }
  // TODO add more tests for resize and read
}
