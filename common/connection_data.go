package common

import (
	"net/rpc"

	"github.com/tauraamui/dragondaemon/pkg/log"
)

func init() {
	if err := rpc.Register(ConnectionData{}); err != nil {
		log.Error("unable to register connection data type for RPC") //nolint
	}
}

type ConnectionData struct {
	UUID,
	Title,
	Size string
}

func (c ConnectionData) GetUUID(args string, dst *string) error {
	*dst = c.UUID
	return nil
}

func (c ConnectionData) GetTitle(args string, dst *string) error {
	*dst = c.Title
	return nil
}

func (c ConnectionData) GetSize(string, dst *string) error {
	*dst = c.Size
	return nil
}
