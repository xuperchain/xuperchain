package mkfile

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
	makeFlags     []string

	*log.Logger
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

// WithMakeFlags set extra flags passing to make command
func (r *Runner) WithMakeFlags(flags []string) *Runner {
	r.makeFlags = flags
	return r
}

// WithLogger set the debug logger
func (r *Runner) WithLogger(logger *log.Logger) *Runner {
	r.Logger = logger
	return r
}

func (r *Runner) mountPaths() []string {
	paths := []string{r.entry.Path, r.xcache}
	if r.xroot != "" {
		paths = append(paths, r.xroot)
	}
	if r.output != "" {
		outdir := filepath.Dir(r.output)
		paths = append(paths, outdir)
	}
	for _, dep := range r.entry.Deps {
		paths = append(paths, dep.Path)
	}
	paths = prefixPaths(paths)
	mounts := make([]string, 0, len(paths)*2)
	for _, path := range paths {
		mounts = append(mounts, "-v", path+":"+path)
	}
	return mounts
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
		"-w", r.entry.Path,
	}
	runargs = append(runargs, mountpaths...)
	runargs = append(runargs, r.image)
	runargs = append(runargs, "emmake", "make", "build", "-f", mkfile)
	runargs = append(runargs, r.makeFlags...)

	r.Printf("docker %s", strings.Join(runargs, " "))
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
	runargs := []string{
		"make", "build", "-f", mkfile,
	}
	runargs = append(runargs, r.makeFlags...)
	r.Printf("emmake %s", strings.Join(runargs, " "))
	cmd := exec.Command("emmake", runargs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// prefixPaths 合并多个路径，剔除有公共前缀的路径，返回所有的最短前缀路径
// 如 /home /lib/a/b /lib/a 最终只会返回 /home和/lib/a
func prefixPaths(paths []string) []string {
	ret := make([]string, 0, len(paths))
	sort.Strings(paths)
	prefix := "\xFF"
	for _, v := range paths {
		if !strings.HasPrefix(v, prefix) {
			ret = append(ret, v)
			prefix = v
		}
	}
	return ret
}
