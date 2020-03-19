package mkfile

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultDockerImage = "hub.baidubce.com/xchain/emcc"
)

type Runner struct {
	entry  *Package
	xroot  string
	xcache string
	image  string
	output string

	withoutDocker bool
}

func NewRunner() *Runner {
	img := os.Getenv("XDEV_CC_IMAGE")
	if img == "" {
		img = defaultDockerImage
	}
	return &Runner{
		image: img,
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

func (r *Runner) WithCacheDir(xcache string) *Runner {
	r.xcache = xcache
	return r
}

func (r *Runner) WithoutDocker() *Runner {
	r.withoutDocker = true
	return r
}

func (r *Runner) WithOutput(out string) *Runner {
	r.output = out
	return r
}

func (r *Runner) mountPaths() []string {
	paths := []string{
		"-v", r.entry.Path + ":/src",
		"-v", r.xcache + ":" + r.xcache,
	}
	if r.xroot != "" {
		paths = append(paths, "-v", r.xroot+":"+r.xroot)
	}
	if r.output != "" {
		outdir := filepath.Dir(r.output)
		paths = append(paths, "-v", outdir+":"+outdir)
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
		"-u", strconv.Itoa(os.Getuid()) + ":" + strconv.Itoa(os.Getgid()),
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
