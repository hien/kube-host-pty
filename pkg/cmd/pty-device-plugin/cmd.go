package ptydp

import (
	"context"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	k8sDP "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"

	"arhat.dev/kube-host-pty/pkg/server"
	"arhat.dev/kube-host-pty/pkg/util"
	"arhat.dev/kube-host-pty/pkg/util/log"
)

const (
	Name = "pty-device-plugin"
)

func NewCmd() (*util.Command, error) {
	var configFile string
	opt := &Options{}
	optFromConfigFile := &Options{}

	cmd := util.DefaultCmd(
		Name, optFromConfigFile, nil,
		func(ctx context.Context, exit context.CancelFunc) error {
			opt.merge(optFromConfigFile)
			return run(ctx, exit, opt)
		},
	)

	cmd.Flags().StringVarP(&opt.KubeletSocket, "kubelet-unix-sock", "k", k8sDP.KubeletSocket, "kubelet service unix sock listening address")
	cmd.Flags().StringVarP(&opt.ListenSocket, "plugin-listen-unix-sock", "l", k8sDP.DevicePluginPath+"arhat.sock", "unix sock address to listen")
	cmd.Flags().StringVarP(&opt.PTSSocketDir, "pts-unix-sock-dir", "d", "/var/run/arhat/pts", "dir to host pts unix sockets")
	cmd.Flags().Uint8VarP(&opt.MaxPtyCount, "max-pty", "m", 10, "maximum pty count allowed on this host")
	cmd.Flags().StringVarP(&opt.Shell, "shell", "s", "sh", "default shell for pty session")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "set config file")

	return cmd, nil
}

func run(ctx context.Context, exit context.CancelFunc, opt *Options) error {
	addressField := log.String("addr", opt.ListenSocket)
	log.D("creating device-plugin service", addressField, log.String("api", k8sDP.Version))

	srv := grpc.NewServer([]grpc.ServerOption{}...)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, unix.SIGQUIT)
	util.Workers.Add(func(func()) (_ interface{}, _ error) {
		for range sigCh {
			srv.GracefulStop()
			exit()
			return
		}
		return
	})

	k8sDP.RegisterDevicePluginServer(srv, server.NewPtyDevicePluginServer(opt.Shell, opt.PTSSocketDir, opt.MaxPtyCount))

	errCh := util.Workers.Add(func(sigContinue func()) (_ interface{}, err error) {
		log.I("ListenAndServe device-plugin", addressField)
		defer log.I("ListenAndServe device-plugin exited", addressField)

		if err = util.GRPCListenAndServe(srv, "unix", opt.ListenSocket); err != nil {
			log.E("ListenAndServe device-plugin failed", addressField)
		}
		return
	})[0].Error

	select {
	case err := <-errCh:
		return err
	default:
		// continue
	}

	conn, err := util.DialGRPC(ctx, "unix", opt.ListenSocket, 5*time.Second, nil)
	if err != nil {
		log.E("dial device-plugin service failed", log.Err(err))
		return err
	} else {
		_ = conn.Close()
	}

	util.InitGraceUpgrade(exit, 30*time.Second, unix.SIGHUP)

	if err := opt.registerResource(ctx); err != nil {
		return err
	}

	return nil
}
