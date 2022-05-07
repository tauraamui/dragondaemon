package rtsp

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/aler9/gortsplib/pkg/base"
)

const clientReadBufferSize = 4096

type Client struct {
	conn       net.Conn
	connReader bufio.Reader
	address    string
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

	c.conn = conn
	c.connReader = *bufio.NewReaderSize(c.conn, clientReadBufferSize)

	return nil
}

func (c *Client) Options() error {
	req := base.Request{
		Method: base.Options,
		URL:    &base.URL{Host: "/"},
	}

	var buff bytes.Buffer
	fmt.Printf("RTSP REQUEST: %s\n", req)
	req.Write(&buff)

	_, err := c.conn.Write(buff.Bytes())
	if err != nil {
		return err
	}

	resp := response{}
	resp.Read(&c.connReader)

	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

type response struct {
	RTSPVersion string
	Status      int
}

func (r *response) Read(rdr io.Reader) {
	scanner := bufio.NewScanner(rdr)
	// first header line will be RTSP version and status code
	if ok := scanner.Scan(); ok {
		line := scanner.Bytes()

		for i, b := range line {
			fmt.Printf("BYTE[%d]: %b, STR: %s\n", i, b, string(b))
		}
	}
}
