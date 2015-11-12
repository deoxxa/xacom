package xacom // import "fknsrs.biz/p/xacom"

import (
	"encoding"
	"fmt"
)

type request interface {
	encoding.TextMarshaler

	messageType() byte
	sequenceNumber() int
	withSequenceNumber(seq int) request
}

type response interface {
	encoding.TextUnmarshaler

	messageType() byte
	sequenceNumber() int
}

func parseResponse(b []byte) (response, error) {
	switch b[0] {
	case 'v':
		var r verifyPagerResponse
		return &r, r.UnmarshalText(b)
	case 'd':
		var r pageDirectResponse
		return &r, r.UnmarshalText(b)
	case 'm':
		var r sendMessageResponse
		return &r, r.UnmarshalText(b)
	}

	return nil, fmt.Errorf("unknown message type %q", b[0])
}

type verifyPagerRequest struct {
	SequenceNumber int
	PagerNumber    string
}

func (v verifyPagerRequest) messageType() byte   { return 'v' }
func (v verifyPagerRequest) sequenceNumber() int { return v.SequenceNumber }
func (v verifyPagerRequest) withSequenceNumber(seq int) request {
	v.SequenceNumber = seq

	return v
}

func (v verifyPagerRequest) MarshalText() ([]byte, error) {
	s := fmt.Sprintf(
		"v%02d%10s\r",
		v.SequenceNumber,
		v.PagerNumber,
	)

	return []byte(s), nil
}

type verifyPagerResponse struct {
	SequenceNumber int
	PagerNumber    string
	PagerStatus    byte
}

func (v verifyPagerResponse) messageType() byte   { return 'v' }
func (v verifyPagerResponse) sequenceNumber() int { return v.SequenceNumber }

func (v *verifyPagerResponse) UnmarshalText(text []byte) error {
	_, err := fmt.Sscanf(
		string(text),
		"v%02d%10s%c",
		&v.SequenceNumber,
		&v.PagerNumber,
		&v.PagerStatus,
	)

	return err
}

type pageDirectRequest struct {
	SequenceNumber  int
	DestinationCode string
	PagerNumber     string
	Message         string
}

func (p pageDirectRequest) messageType() byte   { return 'd' }
func (p pageDirectRequest) sequenceNumber() int { return p.SequenceNumber }
func (p pageDirectRequest) withSequenceNumber(seq int) request {
	p.SequenceNumber = seq

	return p
}

func (p pageDirectRequest) MarshalText() ([]byte, error) {
	s := fmt.Sprintf(
		"d%02d%2s%-10s%s\r",
		p.SequenceNumber,
		p.DestinationCode,
		p.PagerNumber,
		p.Message,
	)

	return []byte(s), nil
}

type pageDirectResponse struct {
	SequenceNumber int
	Status         byte
}

func (p pageDirectResponse) messageType() byte   { return 'd' }
func (p pageDirectResponse) sequenceNumber() int { return p.SequenceNumber }

func (p *pageDirectResponse) UnmarshalText(text []byte) error {
	_, err := fmt.Sscanf(
		string(text),
		"d%02d%c",
		&p.SequenceNumber,
		&p.Status,
	)

	return err
}

type sendMessageRequest struct {
	SequenceNumber int
	PagerNumber    string
	Message        string
}

func (s sendMessageRequest) messageType() byte   { return 'm' }
func (s sendMessageRequest) sequenceNumber() int { return s.SequenceNumber }
func (s sendMessageRequest) withSequenceNumber(seq int) request {
	s.SequenceNumber = seq

	return s
}

func (m sendMessageRequest) MarshalText() ([]byte, error) {
	s := fmt.Sprintf(
		"m%02d%-10s%s\r",
		m.SequenceNumber,
		m.PagerNumber,
		m.Message,
	)

	return []byte(s), nil
}

type sendMessageResponse struct {
	SequenceNumber int
	Status         byte
}

func (s sendMessageResponse) messageType() byte   { return 'm' }
func (s sendMessageResponse) sequenceNumber() int { return s.SequenceNumber }

func (s *sendMessageResponse) UnmarshalText(text []byte) error {
	_, err := fmt.Sscanf(
		string(text),
		"m%02d%c",
		&s.SequenceNumber,
		&s.Status,
	)

	return err
}
