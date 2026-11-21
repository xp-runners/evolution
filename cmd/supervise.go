package cmd

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type supervise struct {
}

// Runs this commandline
func (r supervise) Run(c *commandline) (int, error) {
	stop := make(chan struct{})
	go func() {
		bufio.NewReader(os.Stdin).ReadString('\n')
		close(stop)
	}()

	return superviseRuns(stop, func() (*exec.Cmd, error) { return NewCmd(c) }, 2*time.Second, time.Second)
}

func superviseRuns(stop <-chan struct{}, create func() (*exec.Cmd, error), startup, retry time.Duration) (int, error) {
	for {
		status, elapsed, err := superviseChild(stop, create)
		if err != nil || status == 0 || elapsed < startup {
			if err == nil && status != 0 {
				fmt.Fprintln(os.Stdout, "*** Process exited right after being started, aborting")
			}
			return status, err
		}
		fmt.Fprintf(os.Stdout, "*** Process exited with exitcode %d, respawning...\n", status)
		timer := time.NewTimer(retry)
		select {
		case <-stop:
			timer.Stop()
			return 0, nil
		case <-timer.C:
		}
	}
}

func superviseChild(stop <-chan struct{}, create func() (*exec.Cmd, error)) (int, time.Duration, error) {
	cmd, err := create()
	if err != nil {
		return 2, 0, err
	}
	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		return 2, 0, err
	}
	defer listener.Close()
	cmd.Env = append(cmd.Env, "XP_SIGNAL="+strconv.Itoa(listener.Addr().(*net.TCPAddr).Port))
	cmd.Stdin = nil
	start := time.Now()
	if err := cmd.Start(); err != nil {
		return 2, 0, err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		status, err := exitStatus(err)
		return status, time.Since(start), err
	case <-stop:
		listener.SetDeadline(time.Now().Add(50 * time.Millisecond))
		if conn, err := listener.AcceptTCP(); err == nil {
			conn.SetWriteDeadline(time.Now().Add(time.Second))
			_, err := conn.Write([]byte("QUIT\r\n"))
			conn.Close()
			if err == nil {
				timer := time.NewTimer(2 * time.Second)
				defer timer.Stop()
				select {
				case <-done:
					return 0, time.Since(start), nil
				case <-timer.C:
				}
			}
		}
		_ = cmd.Process.Kill()
		<-done
		return 0, time.Since(start), nil
	}
}

// Name returns this execution model's name
func (r supervise) Name() string {
	return "supervise"
}

// String creates a string representation of this execution model
func (r supervise) String() string {
	return "supervise"
}
