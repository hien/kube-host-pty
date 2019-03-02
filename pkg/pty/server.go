package pty

import (
	"bufio"
	"context"

	"github.com/kr/pty"

	"arhat.dev/kube-host-pty/pkg/util"
	"arhat.dev/kube-host-pty/pkg/util/log"
)

func (t *Terminal) Attach(srv Terminal_AttachServer) error {
	recvCh := make(chan []byte, 1)
	sendCh := make(chan []byte, 1)

	ctx, exit := context.WithCancel(srv.Context())

	util.Workers.Add(func(func()) (interface{}, error) {
		defer close(sendCh)

		// read pty output
		s := bufio.NewScanner(t.ptmx)
		s.Split(util.ScanAnyAvail)

		for s.Scan() {
			ptyOutput := s.Bytes()
			select {
			case <-ctx.Done():
				return nil, nil
			case sendCh <- ptyOutput:
				// continue
			}
		}
		return nil, nil
	}, func(func()) (interface{}, error) {
		defer close(recvCh)

		// read user input
		for {
			inputPacket, err := srv.Recv()
			if err != nil {
				return nil, err
			}

			userInput := inputPacket.GetData()
			select {
			case <-ctx.Done():
				return nil, nil
			case recvCh <- userInput:
				// continue
			}
		}
	})

	defer exit()
	for {
		select {
		case <-ctx.Done():
			return nil
		case ptyOutput, more := <-sendCh:
			if !more {
				return nil
			}

			completed := t.Completed()
			err := srv.Send(&Bytes{Data: ptyOutput, Completed: completed})
			if err != nil {
				log.E("send pty output to user failed", log.Err(err))
				return err
			}

			if completed {
				return nil
			}
		case userInput, more := <-recvCh:
			if !more {
				return nil
			}

			for len(userInput) > 0 {
				n, err := t.ptmx.Write(userInput)
				if err != nil {
					log.E("write user input to pty failed", log.Err(err))
					return err
				}
				userInput = userInput[n:]
			}
		}
	}
}

func (t *Terminal) Resize(ctx context.Context, req *Size) (*Size, error) {
	if err := t.ResizePty(uint16(req.Cols), uint16(req.Rows)); err != nil {
		log.E("resize pty failed", log.Uint32("cols", req.GetCols()), log.Uint32("rows", req.GetRows()), log.Err(err))
		return &Size{}, nil
	}

	// return current pty size
	if rows, cols, err := pty.Getsize(t.ptmx); err != nil {
		log.E("resize pty failed", log.Uint32("cols", req.GetCols()), log.Uint32("rows", req.GetRows()), log.Err(err))
		return &Size{}, nil
	} else {
		return &Size{Rows: uint32(rows), Cols: uint32(cols)}, nil
	}
}
