package mkfile

import (
	"fmt"
	"io"
	"strings"
)

// Makefile represents a Makefile description
type Makefile struct {
	Headers []string
	Tails   []string
	Tasks   []Task
}

// Task represents a single task description in Makefile
type Task struct {
	Target  string
	Deps    []string
	Actions []string
}

// MakeFileWriter write Makefile according to given MakeFile description
type MakeFileWriter struct {
	w io.Writer
}

// NewMakeFileWriter instances a new MakeFileWriter
func NewMakeFileWriter(w io.Writer) *MakeFileWriter {
	return &MakeFileWriter{
		w: w,
	}
}

// Write serializes MakeFile description to MakeFile
func (m *MakeFileWriter) Write(f *Makefile) error {
	m.writeHeaders(f.Headers)
	fmt.Fprintln(m.w)
	for _, task := range f.Tasks {
		m.writeTask(&task)
		fmt.Fprintln(m.w)
	}
	fmt.Fprintln(m.w)
	m.writeTails(f.Tails)
	return nil
}

func (m *MakeFileWriter) writeHeaders(headers []string) error {
	for _, header := range headers {
		fmt.Fprintln(m.w, header)
	}
	return nil
}

func (m *MakeFileWriter) writeTails(tails []string) error {
	for _, tail := range tails {
		fmt.Fprintln(m.w, tail)
	}
	return nil
}

func (m *MakeFileWriter) writeTask(t *Task) error {
	fmt.Fprintf(m.w, "%s: %s\n", t.Target, strings.Join(t.Deps, " "))
	for _, action := range t.Actions {
		fmt.Fprintf(m.w, "\t%s\n", action)
	}
	return nil
}
