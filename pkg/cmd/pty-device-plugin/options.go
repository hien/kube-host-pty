package ptydp

import (
	"context"
	"path/filepath"
	"time"

	k8sDP "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"

	"arhat.dev/kube-host-pty/pkg/constant"
	"arhat.dev/kube-host-pty/pkg/util"
	"arhat.dev/kube-host-pty/pkg/util/log"
)

type Options struct {
	KubeletSocket string `yaml:"kubelet_socket"`
	ListenSocket  string `yaml:"listen_socket"`
	PTSSocketDir  string `yaml:"pts_socket_dir"`
	MaxPtyCount   uint8  `yaml:"max_pty"`
	Shell         string `yaml:"shell"`
}

func (o Options) registerResource(ctx context.Context) error {
	clientConn, err := util.DialGRPC(ctx, "unix", o.KubeletSocket, 5*time.Second, nil)
	if err != nil {
		return err
	}
	defer func() { _ = clientConn.Close() }()

	client := k8sDP.NewRegistrationClient(clientConn)
	endpoint := filepath.Base(o.ListenSocket)

	log.D("start register resource",
		log.String("resource_name", constant.ResourceNamePty),
		log.String("api_version", k8sDP.Version),
		log.String("endpoint", endpoint))

	if _, err = client.Register(ctx, &k8sDP.RegisterRequest{
		Version:      k8sDP.Version,
		Endpoint:     endpoint,
		ResourceName: constant.ResourceNamePty,
	}); err != nil {
		log.E("register resource failed", log.Err(err))
		return err
	}

	return nil
}

func (o *Options) merge(a *Options) {
	if a == nil {
		return
	}

	if a.KubeletSocket != "" {
		o.KubeletSocket = a.KubeletSocket
	}

	if a.ListenSocket != "" {
		o.ListenSocket = a.ListenSocket
	}
}
