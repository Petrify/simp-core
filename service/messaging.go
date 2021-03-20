package service

import (
	"fmt"
	"time"
)

type message struct {
	origin        Service
	target        Service
	data          map[string]interface{}
	responseChan  chan *message
	rTimeout      time.Duration
	sentTime      time.Time
	rTimeoutTimer *time.Timer
	err           error
}

func NewMessage(origin Service) *message {
	m := message{
		origin:       origin,
		responseChan: make(chan *message, 1),
	}

	return &m
}

func (m *message) AddData(key string, val interface{}) {
	m.data[key] = val
}

func (m *message) Send(target int64) (*message, error) {

	m.SendNoWait(target)
	m.rTimeoutTimer = time.NewTimer(m.rTimeout)
	return m.response()
}

func (m *message) SendNoWait(target int64) {
	ts, err := getService(target)
	if err != nil {
		m.err = err
		close(m.responseChan)
		return
	}
	m.target = ts
	m.sentTime = time.Now()
	m.target.giveMessage(m)
}

func (m *message) Respond(src Service) (*message, error) {
	resp := NewMessage(src)
	resp.target = m.origin
	resp.sentTime = time.Now()
	resp.rTimeoutTimer = time.NewTimer(m.rTimeout)
	return resp.response()
}

func (m *message) RespondNoWait(src Service) {
	resp := NewMessage(src)
	resp.target = m.origin
	resp.sentTime = time.Now()
}

func (m *message) ToString() string {
	return "TODO: message.ToString()"
}

func (m *message) SetTimeout(ms int) {
	m.rTimeout = time.Duration(ms) * time.Millisecond
}

func (m *message) response() (*message, error) {

	select {

	// response happens before timeout
	case r, ok := <-m.responseChan:

		if ok {
			//if a response comes through the channel no error is assumed
			return r, nil

		} else {
			// if the channel closes prematurely, the causing error can be found in m.err
			return nil, m.err
		}

	case t := <-m.rTimeoutTimer.C:
		return nil, &timeoutError{t}
	}

}

type timeoutError struct {
	t time.Time
}

func (e *timeoutError) Error() string {
	return fmt.Sprintf("Message response timed out at %s", e.t.String())
}
