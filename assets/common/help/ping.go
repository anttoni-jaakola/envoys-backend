package help

import (
	"net"
	"net/url"
	"strings"
	"time"
)

// The constant 'maxRetries' is used to set a limit to the number of times a program can attempt to execute a particular
// task before giving up and returning an error message. This value can help ensure that the program does not get stuck
// in an infinite loop of retries, which could cause the program to slow down or crash.
const maxRetries = 5

// Ping - The purpose of the function Ping() is to determine whether or not a remote procedure call (RPC) is available. The
// function takes a string argument, which is the address of the RPC, and returns a boolean, indicating whether or not
// the RPC is available.
func Ping(rpc string) bool {

	// This code snippet is used to parse a URL from the string 'rpc' and store the result in the variable u. If an error
	// occurs when parsing the URL, the code returns false.
	u, err := url.Parse(rpc)
	if err != nil {
		return false
	}

	// This code is checking if a given URL is valid. It first checks the scheme of the URL (whether it is "http" or
	// https") and then attempts to establish a connection to the host of the URL. If it is able to establish a connection,
	// the URL is valid, and the code will return true. If it is not able to establish a connection, the code will return false.
	if u.Scheme == "http" || u.Scheme == "https" {
		if conn, err := retryDial(u.Host, time.Second*15); err != nil {
			return false
		} else {
			conn.Close()
			return true
		}
	}

	// This code is checking if the string "rpc" is formatted correctly. The string contains 3 parts separated by a colon
	// (:). The code is using the strings.Split() function to split the string into an array of strings, and then it is
	// checking if the length of the array is 3. If it is not 3, the function returns false.
	addr := strings.Split(rpc, ":")
	if len(addr) != 3 {
		return false
	}

	// The purpose of this code is to check if a connection can be established with a specified host and port, with a given
	// timeout. It uses the retryDial() function to attempt to make the connection, and if successful, closes the connection
	// and returns true. If the connection fails, it returns false.
	if conn, err := retryDial(net.JoinHostPort(addr[1], addr[2]), time.Second*15); err != nil {
		return false
	} else {
		conn.Close()
		return true
	}
}

// retryDial - This function is used to dial a network connection with a given address and timeout. It is designed to retry up to a
// maximum number of times if the initial connection attempt fails. This is useful for establishing a connection to a
// network service that may not be available immediately.
func retryDial(addr string, timeout time.Duration) (net.Conn, error) {

	// The purpose of the code above is to declare two variables, conn and err, of type net.Conn and error respectively.
	// This code allows the user to establish a connection over a network and handle any errors that arise while doing so.
	var (
		conn net.Conn
		err  error
	)

	// The purpose of this code is to set up a loop that will run a maximum of maxRetries times. The loop will start at the
	// value of 0, and increment the value of i by 1 each time the loop runs, and will stop running when i is equal to maxRetries.
	for i := 0; i < maxRetries; i++ {
		time.Sleep(timeout)

		// This code is attempting to establish a network connection with an address (addr) using the TCP protocol, with a
		// timeout of timeout seconds. The connection is stored in the conn variable and any errors are stored in the err
		// variable. If there is no error, the loop is broken.
		conn, err = net.DialTimeout("tcp", addr, timeout)
		if err == nil {
			break
		}
	}

	return conn, err
}
