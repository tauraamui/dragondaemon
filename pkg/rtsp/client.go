package rtsp

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	address string
}

func NewClient(address string) (*Client, error) {
	c := Client{address: address}
	if err := c.validateAndProcessAddr(); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Client) validateAndProcessAddr() error {
	parsedURL, err := url.Parse(c.address)
	if err != nil {
		return err
	}

	if s := parsedURL.Scheme; s != "rtsp" {
		return fmt.Errorf("unsupported scheme: %s ('rtsp' is the only supported scheme)", s)
	}

	c.address = parsedURL.Host
	if !strings.Contains(c.address, ":") {
		c.address += ":554"
	}

	return nil
}

func (c *Client) Connect(cancel context.Context) error {
	var d net.Dialer
	ctx, ccancel := context.WithTimeout(cancel, time.Minute)
	defer ccancel()

	conn, err := d.DialContext(ctx, "tcp", c.address)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("Hello, World!")); err != nil {
		return err
	}

	return nil
}
