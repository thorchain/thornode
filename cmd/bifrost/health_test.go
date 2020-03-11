package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func TestPackage(t *testing.T) { TestingT(t) }

type HealthServerTestSuite struct {
}

var _ = Suite(&HealthServerTestSuite{})

func (HealthServerTestSuite) TestHealthServer(c *C) {
	tssServer := &MockTssServer{}
	s := NewHealthServer("127.0.0.1:8080", tssServer)
	c.Assert(s, NotNil)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := s.Start()
		c.Assert(err, IsNil)
	}()
	time.Sleep(time.Second)
	c.Assert(s.Stop(), IsNil)
}

func (HealthServerTestSuite) TestPingHandler(c *C) {
	tssServer := &MockTssServer{}
	s := NewHealthServer("127.0.0.1:8080", tssServer)
	c.Assert(s, NotNil)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	res := httptest.NewRecorder()
	s.pingHandler(res, req)
	c.Assert(res.Code, Equals, http.StatusOK)
}

func (HealthServerTestSuite) TestGetP2pIDHandler(c *C) {
	tssServer := &MockTssServer{}
	s := NewHealthServer("127.0.0.1:8080", tssServer)
	c.Assert(s, NotNil)
	req := httptest.NewRequest(http.MethodGet, "/p2pid", nil)
	res := httptest.NewRecorder()
	s.getP2pIDHandler(res, req)
	c.Assert(res.Code, Equals, http.StatusOK)
}
