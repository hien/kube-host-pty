package util

import (
	"context"
	"crypto/tls"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/cloudflare/tableflip"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	netOnce = &sync.Once{}
	Net, _  = tableflip.New(tableflip.Options{UpgradeTimeout: tableflip.DefaultUpgradeTimeout})
	N = Net.Fds
)

func InitGraceUpgrade(exit context.CancelFunc, timeout time.Duration, sig os.Signal) {
	netOnce.Do(func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, sig)

		go func() {
			// wait for application upgrade finish
			<-Net.Exit()
			// upgrade done, exit this application
			exit()
			// in case not exit, force exit
			<-time.AfterFunc(timeout, func() { os.Exit(1) }).C
		}()

		go func() {
			for range sigCh {
				err := Net.Upgrade()
				if err != nil {
					continue
				}
			}
		}()

		// finish previous upgrade process
		_ = Net.Ready()
	})
}

func DialGRPC(ctx context.Context, proto, address string, timeout time.Duration, tlsConfig *tls.Config) (*grpc.ClientConn, error) {
	ctx, _ = context.WithTimeout(ctx, timeout)

	options := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return (&net.Dialer{Timeout: timeout}).DialContext(ctx, proto, address)
		}),
	}

	if tlsConfig != nil {
		options = append(options, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		options = append(options, grpc.WithInsecure())
	}

	conn, err := grpc.DialContext(ctx, address, options...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func GRPCListenAndServe(server *grpc.Server, proto, address string) error {
	if proto == "unix" {
		if err := os.MkdirAll(filepath.Dir(address), 0755); err != nil && !os.IsExist(err) {
			return err
		}

		if err := os.Remove(address); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	listen, err := Net.Fds.Listen(proto, address)
	if err != nil {
		return err
	}

	return server.Serve(listen)
}
