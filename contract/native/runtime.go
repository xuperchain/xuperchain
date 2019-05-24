package native

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	units "github.com/docker/go-units"
	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/common/config"
)

const (
	pingTimeoutSecond = 2
)

// Process is the container of running contract
type Process interface {
	// Start 启动Native code进程
	Start() error

	// Stop 停止进程，如果在超时时间内进程没有退出则强制杀死进程
	Stop(timeout time.Duration) error
}

// DockerProcess is the process running as a docker container
type DockerProcess struct {
	basedir       string
	binpath       string
	sockpath      string
	chainSockPath string

	cfg *config.NativeDockerConfig

	id     string
	client *client.Client
	log.Logger
}

func (d *DockerProcess) resourceConfig() (container.Resources, error) {
	var cpuLimit, memLimit int64
	cpuLimit = int64(d.cfg.Cpus * 1e9)
	if d.cfg.Memory != "" {
		var err error
		memLimit, err = units.RAMInBytes(d.cfg.Memory)
		if err != nil {
			return container.Resources{}, err
		}
	}
	return container.Resources{
		NanoCPUs: cpuLimit,
		Memory:   memLimit,
	}, nil
}

// Start implements process interface
func (d *DockerProcess) Start() error {
	volumes := map[string]struct{}{
		d.basedir:       {},
		d.chainSockPath: {},
	}
	gid := strconv.Itoa(os.Getgid())
	cmd := []string{
		d.binpath,
		"--sock", d.sockpath,
		"--chain-sock", d.chainSockPath,
	}

	env := []string{"XCHAIN_UNIXSOCK_GID=" + gid,
		"XCHAIN_PING_TIMEOUT=" + strconv.Itoa(pingTimeoutSecond)}

	user := strconv.Itoa(os.Getuid()) + ":" + strconv.Itoa(os.Getgid())
	config := container.Config{
		AttachStdin:     true,
		AttachStdout:    true,
		AttachStderr:    true,
		OpenStdin:       true,
		Volumes:         volumes,
		Env:             env,
		WorkingDir:      d.basedir,
		NetworkDisabled: true,
		Image:           d.cfg.ImageName,
		Cmd:             cmd,
		User:            user,
	}

	resources, err := d.resourceConfig()
	if err != nil {
		return err
	}
	hostConfig := container.HostConfig{
		AutoRemove: true,
		Resources:  resources,
		Binds: []string{
			d.basedir + ":" + d.basedir,
			d.chainSockPath + ":" + d.chainSockPath,
		},
	}

	ctx := context.TODO()
	resp, err := d.client.ContainerCreate(ctx, &config, &hostConfig, nil, "")
	if err != nil {
		return err
	}
	d.Info("create container success", "id", resp.ID)
	d.id = resp.ID

	err = d.client.ContainerStart(ctx, d.id, types.ContainerStartOptions{})
	if err != nil {
		return err
	}
	d.Info("start container success", "id", resp.ID)
	return nil
}

// Stop implements process interface
func (d *DockerProcess) Stop(timeout time.Duration) error {
	ctx := context.TODO()
	err := d.client.ContainerStop(ctx, d.id, &timeout)
	if err != nil {
		return err
	}
	d.Info("stop container success", "id", d.id)
	statch, errch := d.client.ContainerWait(ctx, d.id, container.WaitConditionNotRunning)
	select {
	case <-statch:
	case <-errch:
	}
	d.Info("wait container success", "id", d.id)
	return nil
}

// HostProcess is the process running as a native process
type HostProcess struct {
	basedir       string
	binpath       string
	sockpath      string
	chainSockPath string

	cmd *exec.Cmd
	log.Logger
}

// Start implements process interface
func (h *HostProcess) Start() error {
	cmd := exec.Command(h.binpath,
		"--sock", h.sockpath,
		"--chain-sock", h.chainSockPath,
	)
	cmd.Dir = h.basedir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
		Pgid:   0,
	}
	cmd.Env = []string{"XCHAIN_PING_TIMEOUT=" + strconv.Itoa(pingTimeoutSecond)}

	if err := cmd.Start(); err != nil {
		return err
	}
	h.Info("start command success", "pid", cmd.Process.Pid)
	h.cmd = cmd
	return nil
}

func processExists(pid int) bool {
	return syscall.Kill(pid, syscall.Signal(0)) == nil
}

// Stop implements process interface
func (h *HostProcess) Stop(timeout time.Duration) error {
	h.cmd.Process.Signal(syscall.SIGTERM)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !processExists(h.cmd.Process.Pid) {
			break
		}
		time.Sleep(time.Second)
	}
	// force kill if timeout
	if !time.Now().Before(deadline) {
		h.cmd.Process.Kill()
	}
	h.Info("stop command success", "pid", h.cmd.Process.Pid)
	return h.cmd.Wait()
}
