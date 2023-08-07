package pbmap

import (
	"errors"

	pb "github.com/mudclimber/relay/internal/proto/hub_agent_protos"
)

type ProtoMap interface {} // for aesthetics purposes for now

type Geometry struct {
  Rows int
  Columns int
}

// TypeMap of mutex unfriendly to basic structs

type Read struct {
  Buf []byte
}

type Login struct {
  Login string
  Geometry Geometry
}

type Resize struct {
  NewSize Geometry
}

func Convert(frame *pb.AgentStreamRequest) (interface{}, error) {
  login := frame.GetLogin()
  if login != nil {
    return Login {
      Login: login.Login,
      Geometry: Geometry {
        Rows: int(login.Rows),
        Columns: int(login.Columns),
      },
    }, nil
  }
  read := frame.GetRead()
  if read != nil {
    return Read {
      Buf: read.Buf,
    }, nil
  }
  resize := frame.GetResize()
  if resize != nil {
    return Resize {
      NewSize: Geometry {
        Rows: int(resize.Rows),
        Columns: int(resize.Columns),
      },
    }, nil
  }

  return nil, errors.New("Unexpected thingy")
}
