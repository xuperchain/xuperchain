package mkfile

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Runner struct {
	entry *Package
	xroot string
	image string

	withoutDocker bool
}

func NewRunner() *Runner {
	return &Runner{
		image: "hub.baidubce.com/xchain/xcompiler-cpp:v3.6",
	}
}

func (r *Runner) WithEntry(pkg *Package) *Runner {
	r.entry = pkg
	return r
}

func (r *Runner) WithXROOT(xroot string) *Runner {
	r.xroot = xroot
	return r
}

func (r *Runner) WithoutDocker() *Runner {
	r.withoutDocker = true
	return r
}

func (r *Runner) mountPaths() []string {
	paths := []string{
		"-v", r.entry.Path + ":/src",
	}
	if r.xroot != "" {
		paths = append(paths, "-v", r.xroot+":"+r.xroot)
	}
	// 对于不在当前package目录下的依赖package，需要mount其根目录
	for _, dep := range r.entry.Deps {
		if strings.HasPrefix(dep.Path, r.entry.Path) {
			continue
		}
		paths = append(paths, "-v", dep.Path+":"+dep.Path)
	}
	return paths
}

func (r *Runner) Make(mkfile string) error {
	if !r.withoutDocker {
		return r.makeUsingDocker(mkfile)
	} else {
		return r.makeUsingHost(mkfile)
	}
}

func (r *Runner) makeUsingDocker(mkfile string) error {
	mountpaths := r.mountPaths()
	runargs := []string{
		"run",
		"-u", strconv.Itoa(os.Getuid()),
		"--rm",
	}
	runargs = append(runargs, mountpaths...)
	runargs = append(runargs, r.image)
	runargs = append(runargs, "emmake", "make", "build", "-f", mkfile)

	cmd := exec.Command("docker", runargs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (r *Runner) makeUsingHost(mkfile string) error {
	cmd := exec.Command("emmake", "make", "build", "-f", mkfile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
