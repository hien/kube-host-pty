package ptycli

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"os/signal"
	"time"

	krPty "github.com/kr/pty"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sys/unix"

	"arhat.dev/kube-host-pty/pkg/constant"
	"arhat.dev/kube-host-pty/pkg/pty"
	"arhat.dev/kube-host-pty/pkg/util"
	"arhat.dev/kube-host-pty/pkg/util/log"
)

const (
	Name = "pty-client"
)

func NewCmd() (*util.Command, error) {
	opt := &Options{}
	cmd := util.DefaultCmd(
		Name, opt, nil,
		func(ctx context.Context, exit context.CancelFunc) error {
			return run(ctx, exit, opt)
		})

	cmd.Flags().StringVarP(&opt.Socket, "sock", "s", "", "set socket to use")

	return cmd, nil
}

func run(ctx context.Context, exit context.CancelFunc, opt *Options) error {
	addr := os.Getenv(constant.EnvironNamePtsUnixSockFile)
	conn, err := util.DialGRPC(ctx, "unix", addr, 5*time.Second, nil)
	if err != nil {
		return err
	}

	log.D("request attach to host pty")
	c := pty.NewTerminalClient(conn)
	client, err := c.Attach(ctx)
	if err != nil {
		log.E("attach host pty failed", log.Err(err))
		return err
	}

	// attached to host pty, prepare stdin for shell
	oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.E("make raw stdin failed", log.Err(err))
	}

	// initial window resize
	resizeRemotePtsForStdin(ctx, c)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, unix.SIGWINCH, os.Interrupt)
	_ = util.Workers.Add(func(func()) (interface{}, error) {
		defer func() {
			exit()
			<-time.AfterFunc(time.Second, func() { os.Exit(1) }).C
		}()

		for {
			select {
			case <-ctx.Done():
				_ = client.CloseSend()
				// recover stdin
				_ = terminal.Restore(int(os.Stdin.Fd()), oldState)
				return nil, nil
			case sig, more := <-sigCh:
				if !more {
					return nil, nil
				}

				switch sig {
				case os.Interrupt:
					return nil, nil
				case unix.SIGWINCH:
					resizeRemotePtsForStdin(ctx, c)
				}
			}
		}
	}, func(func()) (interface{}, error) {
		defer func() {
			_ = os.Stdin.Close()
			exit()
		}()

		// recv host pty output
		for {
			ptyOutput, err := client.Recv()
			if err != nil {
				log.E("recv pts output failed", log.Err(err))
				return nil, err
			}

			output := ptyOutput.GetData()
			_, err = io.Copy(os.Stdout, bytes.NewReader(output))
			if err != nil {
				log.E("copy host pty output to stdout failed", log.Err(err))
				return nil, err
			}

			if ptyOutput.GetCompleted() {
				// remote shell application exited, exit now
				return nil, nil
			}
		}
	}, func(func()) (interface{}, error) {
		defer func() {
			exit()
			_ = os.Stdout.Close()
		}()

		// read and send stdin input
		s := bufio.NewScanner(os.Stdin)
		s.Split(util.ScanAnyAvail)

		for s.Scan() {
			if err := client.Send(&pty.Bytes{Data: s.Bytes()}); err != nil {
				log.E("send user input failed", log.Err(err))
				return nil, err
			}
		}
		return nil, s.Err()
	})

	return nil
}

func resizeRemotePtsForStdin(ctx context.Context, c pty.TerminalClient) {
	rows, cols, err := krPty.Getsize(os.Stdin)
	if err != nil {
		log.E("get stdin pty size failed", log.Err(err))
		return
	}

	hostPtySize, err := c.Resize(ctx, &pty.Size{Cols: uint32(cols), Rows: uint32(rows)})
	if err != nil {
		log.I("resize pty failed", log.Err(err),
			log.Uint32("host_cols", hostPtySize.GetCols()),
			log.Uint32("host_rows", hostPtySize.GetRows()),
			log.Int("stdin_cols", cols),
			log.Int("stdin_rows", rows))
	}
}
