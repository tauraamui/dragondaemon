package common

import "net/rpc"

func init() {
	rpc.Register(ConnectionData{})
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
