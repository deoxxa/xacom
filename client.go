package xacom // import "fknsrs.biz/p/xacom"

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync/atomic"
)

// Client handles the association of requests and responses and generating
// sequence IDs.
type Client struct {
	conn *net.UDPConn
	msgs chan []byte
	errs chan error
	resp [100]chan response
	seq  int64
}

// NewClient creates a new Client object with the specified address.
func NewClient(addr string) (*Client, error) {
	raddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp4", nil, raddr)
	if err != nil {
		return nil, err
	}

	c := Client{
		conn: conn,
		msgs: make(chan []byte, 10),
		errs: make(chan error, 1),
		seq:  rand.Int63(),
	}

	return &c, nil
}

// Run is the main loop for the client - it must be running for the client to
// operate. Usually one would run it like so:
//
// `go c.Run()`
func (c *Client) Run() error {
	go func() {
		for {
			buf := make([]byte, 1024)

			n, a, err := c.conn.ReadFrom(buf)
			if err == io.EOF {
				c.errs <- nil
				return
			} else if err != nil {
				c.errs <- err
				return
			}

			if a.String() != c.conn.RemoteAddr().String() {
				continue
			}

			c.msgs <- buf[0:n]
		}
	}()

	for {
		select {
		case err := <-c.errs:
			for i, ch := range c.resp {
				if ch != nil {
					ch <- nil
				}

				c.resp[i] = nil
			}

			return err
		case buf := <-c.msgs:
			r, err := parseResponse(buf)
			if err != nil {
				return err
			}

			w := c.resp[r.sequenceNumber()]
			if w == nil {
				continue
			}

			c.resp[r.sequenceNumber()] = nil

			w <- r
		}
	}
}

// Close destroys the underlying connection for the client, causing the main
// loop to stop.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Verify checks the status of a pager using the "v" command. The byte returned
// will be one of the following:
//
// N: No such pager
// I: In service
// O: Out of service
func (c *Client) Verify(number string) (byte, error) {
	res, err := c.do(verifyPagerRequest{
		PagerNumber: number,
	})

	if err != nil {
		return 0, err
	}

	if res, ok := res.(*verifyPagerResponse); ok {
		return res.PagerStatus, nil
	}

	return 0, fmt.Errorf("invalid response type; expected %T but got %T", &verifyPagerResponse{}, res)
}

// SendMessage sends a message to a pager using the deprecated "m" command. The
// boolean value returned represents the success of queueing the message for
// delivery - notably this does _not_ mean that the message made it to the
// pager, only to the queue to be sent (maybe later on).
func (c *Client) SendMessage(number, message string) (bool, error) {
	res, err := c.do(sendMessageRequest{
		PagerNumber: number,
		Message:     message,
	})

	if err != nil {
		return false, err
	}

	if res, ok := res.(*sendMessageResponse); ok {
		return res.Status == 0x06, nil
	}

	return false, fmt.Errorf("invalid response type; expected %T but got %T", &verifyPagerResponse{}, res)
}

func (c *Client) do(req request) (response, error) {
	seq := int(atomic.AddInt64(&c.seq, 1) % 100)

	d, err := req.withSequenceNumber(seq).MarshalText()
	if err != nil {
		return nil, err
	}

	c.resp[seq] = make(chan response, 1)

	if _, err := c.conn.Write(d); err != nil {
		return nil, err
	}

	res := <-c.resp[seq]
	if res == nil {
		return nil, fmt.Errorf("got no response")
	}

	return res, nil
}
