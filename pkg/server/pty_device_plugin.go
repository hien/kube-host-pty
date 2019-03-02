package server

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	k8sDP "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"

	"arhat.dev/kube-host-pty/pkg/constant"
	"arhat.dev/kube-host-pty/pkg/pty"
	"arhat.dev/kube-host-pty/pkg/util"
	"arhat.dev/kube-host-pty/pkg/util/log"
)

func NewPtyDevicePluginServer(shell, rootDir string, maxPty uint8) k8sDP.DevicePluginServer {
	return &devicePluginService{
		shell:  shell,
		ptyDir: rootDir,
		devices: func() []*k8sDP.Device {
			// generate virtual pty devices
			var devices []*k8sDP.Device
			for i := uint8(0); i < maxPty; i++ {
				id := fmt.Sprintf("pts%d", i)
				devices = append(devices, &k8sDP.Device{ID: id, Health: k8sDP.Healthy})
			}
			return devices
		}(),
	}
}

type devicePluginService struct {
	shell            string
	ptyDir           string
	devices          []*k8sDP.Device
	allocatedDevices sync.Map
}

// GetDevicePluginOptions returns
func (*devicePluginService) GetDevicePluginOptions(context.Context, *k8sDP.Empty) (*k8sDP.DevicePluginOptions, error) {
	return &k8sDP.DevicePluginOptions{PreStartRequired: false}, nil
}

// ListAndWatch returns all pty devices available
func (svc *devicePluginService) ListAndWatch(_ *k8sDP.Empty, srv k8sDP.DevicePlugin_ListAndWatchServer) error {
	if err := srv.Send(&k8sDP.ListAndWatchResponse{Devices: svc.devices}); err != nil {
		return err
	}

	select {
	case <-srv.Context().Done():
		return nil
	}
}

// Allocate a new pty session before container creation
// Every pod with pty request can only have one pty device allocated
// so, you SHOULD NOT request more than one pty in your container
func (svc *devicePluginService) Allocate(ctx context.Context, req *k8sDP.AllocateRequest) (*k8sDP.AllocateResponse, error) {
	containerResp := make([]*k8sDP.ContainerAllocateResponse, 0)

	for _, r := range req.GetContainerRequests() {
		log.D("allocate pty device", log.Strings("devicesIDs", r.GetDevicesIDs()))
		devIDs := r.GetDevicesIDs()
		if len(devIDs) == 0 {
			return nil, fmt.Errorf("no dev id provided")
		}

		var (
			pseudoID        = devIDs[0]
			hostPtySockDir  = filepath.Join(svc.ptyDir, pseudoID)
			hostPtsSockFile = filepath.Join(hostPtySockDir, pseudoID)
			ctrPtySockDir   = "/var/run/arhat/pts/"
			ctrPtsSockFile  = filepath.Join(ctrPtySockDir, pseudoID)
		)

		// NEVER try to reuse previous pts session, close if still active
		// move this to Deallocate call if possible
		// see https://github.com/kubernetes/kubernetes/issues/59110 for related discussion
		if val, ok := svc.allocatedDevices.Load(pseudoID); ok {
			term := val.(*pty.Terminal)
			if !term.Completed() {
				_ = term.Close()
			}
			svc.allocatedDevices.Delete(pseudoID)
		}

		// always allocate new pty session
		log.D("open host pty for device allocation")
		term, err := pty.Open(svc.shell, 80, 30)
		if err != nil {
			log.E("create terminal pts failed", log.Err(err))
			return nil, fmt.Errorf("create terminal pts failed")
		}

		util.Workers.Add(func(func()) (_ interface{}, err error) {
			addressField := log.String("addr", hostPtsSockFile)

			log.I("ListenAndServe pts", addressField)
			defer log.I("ListenAndServe pts exited", addressField)

			if err = term.ListenAndServe(hostPtsSockFile); err != nil {
				log.E("ListenAndServe pts failed", addressField)
			}
			return
		})

		// wait for server with a self dial
		conn, err := util.DialGRPC(ctx, "unix", hostPtsSockFile, 5*time.Second, nil)
		if err != nil {
			// can't dial to the pts sock destroy this pty and its services
			log.E("dial pts service failed", log.Err(err))
			_ = term.Close()
			return nil, err
		} else {
			_ = conn.Close()
		}

		svc.allocatedDevices.Store(pseudoID, term)
		containerResp = append(containerResp, &k8sDP.ContainerAllocateResponse{
			Envs:   map[string]string{constant.EnvironNamePtsUnixSockFile: ctrPtsSockFile},
			Mounts: []*k8sDP.Mount{{ContainerPath: ctrPtySockDir, HostPath: hostPtySockDir}},
		})
	}

	return &k8sDP.AllocateResponse{ContainerResponses: containerResp}, nil
}

// PreStartContainer is called, if indicated by Device Plugin during registration phase,
// before each container start. Device devicePlugin can run device specific operations
// such as resetting the device before making devices available to the container
func (*devicePluginService) PreStartContainer(ctx context.Context, req *k8sDP.PreStartContainerRequest) (*k8sDP.PreStartContainerResponse, error) {
	return &k8sDP.PreStartContainerResponse{}, nil
}
