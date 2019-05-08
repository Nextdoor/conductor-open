package datadog

import (
	"net"
	"os"
	"testing"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/stretchr/testify/assert"
)

var addr, _ = net.ResolveUDPAddr("udp", "localhost:8125")
var sock, _ = net.ListenUDP("udp", addr)

func TestLogInfo(t *testing.T) {
	os.Setenv("STATSD_HOST", "localhost")
	enableDatadog = true
	c = newStatsdClient()
	assert.NotNil(t, c)
	log(statsd.Info, "%s testing", "conductor")
	buf := make([]byte, 1024)
	rlen, _, _ := sock.ReadFromUDP(buf)
	assert.Contains(t, string(buf[0:rlen]), "conductor testing")
}

func TestLogError(t *testing.T) {
	os.Setenv("STATSD_HOST", "localhost")
	enableDatadog = true
	c = newStatsdClient()
	assert.NotNil(t, c)
	log(statsd.Error, "%s testing", "conductor")
	buf := make([]byte, 1024)
	rlen, _, _ := sock.ReadFromUDP(buf)
	assert.Contains(t, string(buf[0:rlen]), "conductor testing")
}
