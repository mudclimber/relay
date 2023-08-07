package draw

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"log"

	"github.com/mudclimber/relay/internal/pipes/pbmap"
)

var CSI = "\033["
var CLEAR = "\033[1;1H\033[J"

func pos(row int, column int) string {
	return CSI + strconv.FormatInt(int64(row), 10) + ";" + strconv.FormatInt(int64(column), 10) + "H"
}

func forward(n int) string {
	return CSI + strconv.FormatInt(int64(n), 10) + "C"
}

type Box struct {
	Geometry  pbmap.Geometry
	drawEdges []int
  PromptBuffer []byte
}

func (b Box) DrawBox() []byte {
	buf := bytes.Buffer{}

	cols := b.Geometry.Columns
	rows := b.Geometry.Rows

	buf.WriteString(CLEAR + "┏")
	buf.WriteString(strings.Repeat("━", cols-2))
	buf.WriteString("┓\r\n")

	for l := 0; l < rows-4; l++ {
		buf.WriteString("┃" + forward(cols-2) + "┃\r\n")
	}

	// border between game output and prompt
	buf.WriteString("┣")
	buf.WriteString(strings.Repeat("━", cols-2))
	buf.WriteString("┫\r\n")

	// prompt line
	buf.WriteString("┃" + forward(cols-2) + "┃\r\n")

	// border under prompt
	buf.WriteString("┗")
	buf.WriteString(strings.Repeat("━", cols-2))
	buf.WriteString("┛")

	return buf.Bytes()
}

func (b Box) DrawPrompt() []byte {
  prompt := make([]byte, len(b.PromptBuffer))
  copy(prompt, b.PromptBuffer)
  if len(prompt) > b.Geometry.Columns - 2 {
    prompt = prompt[len(prompt) - (b.Geometry.Columns - 3):]
  }
  return []byte(
    pos(b.Geometry.Rows - 1, 2) +
    string(prompt) +
    strings.Repeat(" ", b.Geometry.Columns - 2 - len(prompt)) +
    pos(b.Geometry.Rows - 1, 2 + len(prompt)))
}

func (b Box) GotoPrompt() []byte {
  return []byte(pos(b.Geometry.Rows - 1, 2 + len(b.PromptBuffer)))
}

func (b Box) DrawLines(lines [][]byte) ([]byte, error) {
	if len(lines) > b.Geometry.Rows-4 {
		return nil, errors.New("Rows wrong size")
	}

	if len(lines) == 0 {
		return nil, errors.New("toooooo small")
	}

	if len(lines[0]) > b.Geometry.Columns-2 {
		return nil, errors.New("Columns wrong size")
	}
	buf := bytes.Buffer{}
	row := b.Geometry.Rows - 4
	for i := len(lines) - 1; i >= 0; i-- {
    if len(lines[i]) > b.Geometry.Columns - 2 {
      log.Printf(
        "line exceeded. %d -> %d\n",
        i, len(lines[i]))
      lines[i] = lines[i][:b.Geometry.Columns - 2]
    }
		buf.WriteString(pos(row, 2))
		row--
    buf.Write(lines[i])
    buf.Write(bytes.Repeat([]byte(" "), b.Geometry.Columns - 2 - len(lines[i])))
	}
	return buf.Bytes(), nil
}
