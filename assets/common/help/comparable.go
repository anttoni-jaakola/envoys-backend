package help

import (
	"encoding/json"
	"github.com/go-ping/ping"
	"net"
	"net/url"
	"strings"
	"time"
)

func Comparable(bytea []byte, index string, addColumn ...string) bool {

	var (
		array []string
	)

	if err := json.Unmarshal(bytea, &array); err != nil {
		return false
	}

	array = append(array, addColumn...)
	if IndexOf(array, index) {
		return true
	}

	return false
}

func IndexOf[T comparable](collection []T, el T) bool {
	for _, x := range collection {
		if x == el {
			return true
		}
	}
	return false
}

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
