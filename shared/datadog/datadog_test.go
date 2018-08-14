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
	defer func() {
		buf := make([]byte, 1024)
		rlen, _, _ := sock.ReadFromUDP(buf)
		assert.Contains(t, string(buf[0:rlen]), "conductor testing")
	}()
	os.Setenv("STATSD_HOST", "localhost")
	assert.NotNil(t, c)
	log(statsd.Info, "%s testing", "conductor")
}

func TestLogError(t *testing.T) {
	defer func() {
		buf := make([]byte, 1024)
		rlen, _, _ := sock.ReadFromUDP(buf)
		assert.Contains(t, string(buf[0:rlen]), "conductor testing")
	}()
	os.Setenv("STATSD_HOST", "localhost")
	assert.NotNil(t, c)
	log(statsd.Error, "%s testing", "conductor")
}
