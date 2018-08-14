package datadog

import (
	"net"
	"os"
	"testing"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	addr, _ := net.ResolveUDPAddr("udp", "localhost:8125")
	sock, _ := net.ListenUDP("udp", addr)
	defer func() {
		buf := make([]byte, 1024)
		rlen, _, _ := sock.ReadFromUDP(buf)
		assert.Contains(t, string(buf[0:rlen]), "conductor testing")
	}()
	os.Setenv("STATSD_HOST", "localhost")
	assert.NotNil(t, c)
	log(statsd.Info, "%s testing", "conductor")
}
