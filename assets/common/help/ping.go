package help

import (
	"github.com/go-ping/ping"
	"net"
	"net/url"
	"strings"
	"time"
)

// Ping - rpc/api ping status.
func Ping(rpc string) bool {

	addr := strings.Split(rpc, ":")
	if len(addr) != 3 {

		parse, err := url.Parse(rpc)
		if err != nil {
			return false
		}

		stats, err := ping.NewPinger(parse.Host)
		if err != nil {
			return false
		}

		stats.Timeout = 2
		stats.Count = 1

		err = stats.Run()
		if err != nil {
			return false
		}

		return true
	}

	if strings.Contains(addr[1], "//") {

		conn, err := net.DialTimeout("tcp", net.JoinHostPort(strings.ReplaceAll(addr[1], "//", ""), addr[2]), time.Second)
		if err != nil {
			return false
		}
		if conn == nil {
			return false
		}
		defer conn.Close()

		return true
	}

	return false
}
