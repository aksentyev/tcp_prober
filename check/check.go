package check

import (
    "net"
    "time"
    "fmt"
    "io"
)

// Test port is open. Note that UDP ports are acceptable forever.
func Poke(addr, port string) int {
    conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", addr, port))
    if err != nil {
        return 1
    }

    defer conn.Close()
    conn.SetReadDeadline(time.Now().Add(1 * time.Second))
    buf := make([]byte, 0)

    if _, err = conn.Read(buf); err == io.EOF {
        return 1
    }

    return 0
}
