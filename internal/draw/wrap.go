package draw

import (
	"log"

	"github.com/mudclimber/relay/internal/pipes/pbmap"
)

type BoxAdapter struct {
	History [][]byte
}

func (ba BoxAdapter) Foo() {
  log.Panic("oh no")
}

func (ba BoxAdapter) wrapAppend(outputGeom pbmap.Geometry) [][]byte {
	wrapped := [][]byte{}
	for i := len(ba.History) - 1; i >= 0; i-- {
		if len(ba.History[i]) <= outputGeom.Columns {
			wrapped = append(wrapped, ba.History[i])
			if len(wrapped) == outputGeom.Rows {
				return wrapped
			}
		}
		line := ba.History[i]
		if len(line) > outputGeom.Columns {
			remainderSize := len(line) % outputGeom.Columns
			lastLine := line[len(line)-remainderSize:]
			wrapped = append(wrapped, lastLine)
			if len(wrapped) == outputGeom.Rows {
				return wrapped
			}

			iter := len(line) - remainderSize - outputGeom.Columns
			// loop through in <col> size chunks
			for iter >= 0 {
				nextLine := line[iter : iter + outputGeom.Columns]
				wrapped = append(wrapped, nextLine)
				if len(wrapped) == outputGeom.Rows {
					return wrapped
				}
				iter -= outputGeom.Columns
			}
			// Get the last part the loop couldn't get
			if iter+outputGeom.Columns > 0 {
				wrapped = append(wrapped, line[:iter+outputGeom.Columns])
			}
		}
	}
	return wrapped
}

func (ba BoxAdapter) Wrap(outputGeom pbmap.Geometry) [][]byte {
	wrapped := ba.wrapAppend(outputGeom)
	for i, j := 0, len(wrapped)-1; i < j; i, j = i+1, j-1 {
		wrapped[i], wrapped[j] = wrapped[j], wrapped[i]
	}
	return wrapped
}
