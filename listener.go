package main

import (
	"fmt"
	"net"
)

type listener struct {
	net.Listener
	ch chan struct{}
}

type conn struct {
	net.Conn
	closeCallback func()
}

func New(network, addr string, maxConcurrentRequests int) (net.Listener, error) {
	l, err := net.Listen(network, addr)
	if err != nil {
		return nil, fmt.Errorf("net.Listen: %w", err)
	}

	return &listener{
		Listener: l,
		ch:       make(chan struct{}, maxConcurrentRequests),
	}, nil
}

func (l *listener) Accept() (net.Conn, error) {
	l.ch <- struct{}{}

	c, err := l.Listener.Accept()
	if err != nil {
		<-l.ch
		return nil, fmt.Errorf("Listener.Accept: %w", err)
	}

	return &conn{
		Conn:          c,
		closeCallback: func() { <-l.ch },
	}, nil
}

func (c *conn) Close() error {
	defer c.closeCallback()
	return c.Conn.Close()
}
