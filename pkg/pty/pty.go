package pty

import (
	"os"
	"os/exec"
	"runtime"
	"sync/atomic"

	"github.com/kr/pty"
	"google.golang.org/grpc"

	"arhat.dev/kube-host-pty/pkg/util"
)

const (
	defaultUnixShell    = "sh"
	defaultWindowsShell = "cmd.exe"
)

type Terminal struct {
	ptmx      *os.File
	cmd       *exec.Cmd
	completed uint32
}

func (t *Terminal) Completed() bool {
	return atomic.LoadUint32(&t.completed) == 1
}

func (t *Terminal) ResizePty(cols, rows uint16) error {
	return pty.Setsize(t.ptmx, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
}

func (t *Terminal) Close() error {
	if !t.Completed() {
		_ = t.cmd.Process.Kill()
		return t.ptmx.Close()
	}

	return nil
}

func (t *Terminal) ListenAndServe(addr string) error {
	srv := grpc.NewServer([]grpc.ServerOption{}...)
	RegisterTerminalServer(srv, t)

	return util.GRPCListenAndServe(srv, "unix", addr)
}

func Open(shell string, cols, rows uint16) (*Terminal, error) {
	if shell == "" {
		switch runtime.GOOS {
		case "windows":
			shell = defaultWindowsShell
		default:
			shell = defaultUnixShell
		}
	}

	cmd := exec.Command(shell)
	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: cols, Rows: rows})
	if err != nil {
		return nil, err
	}

	term := &Terminal{ptmx: ptmx, cmd: cmd}
	go func() {
		_ = cmd.Wait()
		atomic.StoreUint32(&term.completed, 1)
		_ = ptmx.Close()
	}()

	return term, nil
}
