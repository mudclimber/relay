package pipes

import (
	"bytes"
	"log"

	"github.com/mudclimber/relay/internal/draw"
	"github.com/mudclimber/relay/internal/pipes/pbmap"
	"github.com/mudclimber/relay/internal/pipes/terminalops"
	"github.com/mudclimber/relay/pkg/agents"
	"github.com/mudclimber/relay/pkg/sockets"
)

var CSI = "\033["
var CLEAR = "\033[1;1H\033[J"

type TerminalProcessor struct {
  Terminal <-chan terminalops.TerminalOp
  ToAgent chan<- agents.ToAgent
  FromSocket <-chan sockets.FromSocket
  ToSocket chan<- sockets.ToSocket
  GameInitialize chan<- terminalops.TerminalOp
  HandlerInitDone chan []byte // receives the intro message
  TermSize pbmap.Geometry
  Box draw.Box
  login string
  promptBuffer []byte
  // LOTS of optimization to be done
  outputBuffer bytes.Buffer
  introFromInit []byte

  // terminal <-chan
  // -> relay bytes to agent client sink
  // game bytes <-chan
  // -> relayed to EvenniaSink(or processed here)
}

/*
NOTE:

When we get fromGame input, update lines,
send line updates in ANSI to agent
*/
func (t *TerminalProcessor) processInput(buf []byte) {
  // TODO: collapse changes into simplest operations before processing
  for _, b := range buf {
    if b == 127 {
      if len(t.promptBuffer) > 0 {
        t.promptBuffer = t.promptBuffer[:len(t.promptBuffer) - 1]
      }
    } else if b >= 32 && b <= 126 {
      t.promptBuffer = append(t.promptBuffer, b)
    } else if b == 13 {
      t.ToSocket <- sockets.SocketInput {
        Buf: append(t.promptBuffer, 13, 10),
      }
      t.promptBuffer = []byte{}
    }
  }
}

// Will maybe want to use this later.
func stuff(inBuf []byte, n int) {
      for n < len(inBuf) && inBuf[n] != 0 {
        //skipped := false
        if inBuf[n] == 255 && n + 2 < len(inBuf) {
          n += 3
          continue
        }
        if inBuf[n] == 27 {
          offset := 1
          for m := 0; m < 3; m++ {
            switch m {
            case 0:
              if n + offset < len(inBuf) {
                if inBuf[n + offset] == '[' {
                  offset++
                  continue
                } else {
                  break
                }
              }
            case 1:
              for (n + offset < len(inBuf) && (inBuf[n + offset] >= '0' && inBuf[n + offset] <= '9') || inBuf[n + offset] == ';') {
                offset++
              }
            case 2:
              if n + offset < len(inBuf) {
              }
              if n + offset < len(inBuf) && inBuf[n + offset] == 'm' {
                n += offset
                //skipped = true
              }
            }
          }
          continue
        }
        // newBuf = append(newBuf, inBuf[n])
        n++
      }
}
var BUFFER_CAP = 80 * 24 * 20

func (t *TerminalProcessor) writeToAgent(buf []byte) {
  t.ToAgent <- agents.AgentWrite {
    Buf: buf,
  }
}

func (t *TerminalProcessor) writeStringToAgent(msg string) {
  t.ToAgent <- agents.AgentWrite {
    Buf: []byte(msg),
  }
}

func (t *TerminalProcessor) redrawAll() {
  boxAdapter := draw.BoxAdapter {
    History: bytes.Split(t.outputBuffer.Bytes(), []byte{13, 10}),
  }
  wrap := boxAdapter.Wrap(pbmap.Geometry {
    Rows: t.TermSize.Rows - 4,
    Columns: t.TermSize.Columns - 2,
  })
  t.writeStringToAgent(CLEAR)
  t.writeToAgent(t.Box.DrawBox())
  // for _, w := range wrap {
  //   log.Println("[wrap]", w)
  // }
  lines, err := t.Box.DrawLines(wrap)
  if err != nil {
    log.Println(err)
    return
  }
  t.writeToAgent(lines)
  // need to pick one var and stick with it
  t.Box.PromptBuffer = t.promptBuffer
  t.writeToAgent(t.Box.DrawPrompt())
}

// Waits for init frame from agent, relays
// to game sink, then waits for game sink to report
// back before processing terminal stuff
func (t *TerminalProcessor) waitForInit() {
  Init:
    for {
      select {
      case op := <-t.Terminal:
        switch op.(type) {
        case terminalops.Initialize:
          initOp := op.(terminalops.Initialize)
          t.TermSize = initOp.Size
          t.login = initOp.Login
          t.Box = draw.Box {
            Geometry: initOp.Size,
          }
          t.writeStringToAgent(CLEAR)
          t.writeToAgent(t.Box.DrawBox())
          t.GameInitialize <- initOp
          t.writeToAgent(t.Box.DrawPrompt())
          break Init
        }
      }
    }
  // Wait for game sink to send the init done from its side
  t.introFromInit = <-t.HandlerInitDone
}

func (t *TerminalProcessor) handleSocketOutput(inBuf []byte) {
  newBuf := []byte{}
  n := 0

  for n < len(inBuf) && inBuf[n] != 0 {
    //skipped := false
    if inBuf[n] == 255 && n + 2 < len(inBuf) {
      n += 3
      continue
    }
    if inBuf[n] == 255 && n == len(inBuf) - 2 {
      n += 2
      // some of the messages in in [255 241], this skips those
      continue
    }
    if inBuf[n] == 27 && n + 1 < len(inBuf) && inBuf[n + 1] != '[' {
      // only allow CSI commands
      newBuf = append(newBuf, []byte("(esc)")...)
      n++
    }
    newBuf = append(newBuf, inBuf[n])
    n++
  }
  // Skip telnet commands in a bad way
  // TODO: manage in a building buf once all is working
  t.outputBuffer.Write(newBuf)
  if t.outputBuffer.Len() > 20 {
    log.Println("[buf tail]", t.outputBuffer.Bytes()[t.outputBuffer.Len() - 17:])
    log.Println("[buf head]", t.outputBuffer.Bytes()[:20])
  }
  if t.outputBuffer.Len() > BUFFER_CAP {
    a := make([]byte, BUFFER_CAP)
    copy(a, t.outputBuffer.Bytes()[t.outputBuffer.Len() - BUFFER_CAP:])
  }
  // todo: write buffer into BoxAdapter in a future method I make and
  // store the history there
  boxAdapter := draw.BoxAdapter {
    History: bytes.Split(t.outputBuffer.Bytes(), []byte{13, 10}),
  }
  wrap := boxAdapter.Wrap(pbmap.Geometry {
    Rows: t.TermSize.Rows - 4,
    Columns: t.TermSize.Columns - 2,
  })
  t.writeToAgent(t.Box.DrawBox())
  // for _, w := range wrap {
  //   log.Println("[wrap]", w)
  // }
  lines, err := t.Box.DrawLines(wrap)
  if err != nil {
    log.Println(err)
    return
  }
  t.writeToAgent(lines)
  t.writeToAgent(t.Box.DrawPrompt())
}

func (t *TerminalProcessor) Run() {
  t.waitForInit()
  t.writeStringToAgent(CLEAR)
  t.handleSocketOutput(t.introFromInit)
  for {
    select {
    case op := <-t.Terminal:
      switch op.(type) {
      case terminalops.Clear:
        t.writeStringToAgent(CLEAR)
      case terminalops.NewGeometry:
        t.TermSize = op.(terminalops.NewGeometry).Size
        t.Box.Geometry = t.TermSize
      case terminalops.RedrawBox:
        t.redrawAll()
      case terminalops.AgentInput:
        t.processInput(op.(terminalops.AgentInput).Buf)
        t.Box.PromptBuffer = t.promptBuffer
        //t.Agent <- WriteResponse {
        //  Buf: t.Box.GotoPrompt(),
        //}
        t.writeToAgent(t.Box.DrawPrompt())
      case terminalops.ReplacePrompt:
        t.Box.PromptBuffer = op.(terminalops.ReplacePrompt).NewPrompt
        t.writeToAgent(t.Box.DrawPrompt())
      default:
        log.Println("[Terminal] Unknown choice encountered")
      }
    case gameMsg := <-t.FromSocket:
      // handle lines here
      switch gameMsg.(type) {
      case sockets.SocketOutput:
        inBuf := gameMsg.(sockets.SocketOutput).Buf
        t.handleSocketOutput(inBuf)
      case agents.AgentClose:
        t.ToAgent <- agents.AgentEOF{}
      }
    }
  }
}
