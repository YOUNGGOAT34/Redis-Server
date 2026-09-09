// Command healthcheck is a tiny, standalone binary used only as a Docker
// HEALTHCHECK for the CacheDB image. It exists because the runtime image is
// distroless (no shell, no coreutils, no redis-cli) so a shelled-out
// `nc`/`redis-cli PING` isn't available. It does exactly one thing: open a
// TCP connection to the given address, send a RESP PING, and check for a
// PONG reply - the same non-mutating readiness signal a real client would
// use, with no effect on server state.
//
// This is intentionally its own package, separate from the main CacheDB
// binary and its command dispatch: it does not import or modify any
// existing application logic.
package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	addr := "127.0.0.1:6379"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck: dial failed:", err)
		os.Exit(1)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(2 * time.Second))

	if _, err := conn.Write([]byte("*1\r\n$4\r\nPING\r\n")); err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck: write failed:", err)
		os.Exit(1)
	}

	buf := make([]byte, 7)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck: read failed:", err)
		os.Exit(1)
	}

	if string(buf[:n]) != "+PONG\r\n" {
		fmt.Fprintf(os.Stderr, "healthcheck: unexpected response: %q\n", buf[:n])
		os.Exit(1)
	}

	os.Exit(0)
}
