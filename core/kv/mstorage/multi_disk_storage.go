// Copyright (c) 2012, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
//
// ~~ 2018.07.11
// Modified by Baidu,Inc. in order to support multi disks storage

package mstorage

import (
	"errors"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"io"
	"io/ioutil"
	"os"
	pt "path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	sstFormat = "%06d.ldb"
)

var (
	errFileOpen = errors.New("leveldb/mstorage: file still open")
	errReadOnly = errors.New("leveldb/mstorage: storage is read-only")
)

func isCorrupted(err error) bool {
	switch err.(type) {
	case *storage.ErrCorrupted:
		return true
	}
	return false
}

type fileLock interface {
	release() error
}

// MultiDiskStorageLock data structure consists of pointer to MultiDiskStorage
type MultiDiskStorageLock struct {
	fs *MultiDiskStorage
}

// Unlock keep instance of MultiDiskStorageLock available
func (lock *MultiDiskStorageLock) Unlock() {
	if lock.fs != nil {
		lock.fs.mu.Lock()
		defer lock.fs.mu.Unlock()
		if lock.fs.slock == lock {
			lock.fs.slock = nil
		}
	}
}

type int64Slice []int64

func (p int64Slice) Len() int           { return len(p) }
func (p int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func writeFileSynced(filename string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Sync(); err == nil {
		err = err1
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

const logSizeThreshold = 1024 * 1024 // 1 MiB

// MultiDiskStorage is a file-system backed storage.
type MultiDiskStorage struct {
	path      string
	dataPaths []string
	readOnly  bool

	mu      sync.Mutex
	flock   fileLock
	slock   *MultiDiskStorageLock
	logw    *os.File
	logSize int64
	buf     []byte
	// Opened file counter; if open < 0 means closed.
	open int
	day  int
}

// OpenFile returns instance of MultiDiskStorage supporting multi disks
func OpenFile(path string, readOnly bool, dataPaths []string) (storage.Storage, error) {
	if fi, err := os.Stat(path); err == nil {
		if !fi.IsDir() {
			return nil, fmt.Errorf("leveldb/storage: open %s: not a directory", path)
		}
	} else if os.IsNotExist(err) && !readOnly {
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	flock, err := newFileLock(filepath.Join(path, "LOCK"), readOnly)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			flock.release()
		}
	}()

	var (
		logw    *os.File
		logSize int64
	)
	if !readOnly {
		logw, err = os.OpenFile(filepath.Join(path, "LOG"), os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		logSize, err = logw.Seek(0, os.SEEK_END)
		if err != nil {
			logw.Close()
			return nil, err
		}
	}
	fullDataPaths := []string{}
	for _, dp := range dataPaths {
		fullPath := pt.Join(dp, pt.Base(path))
		fullDataPaths = append(fullDataPaths, fullPath)
		err := os.MkdirAll(fullPath, os.FileMode(0755))
		if err != nil {
			return nil, err
		}
	}
	fs := &MultiDiskStorage{
		path:      path,
		readOnly:  readOnly,
		flock:     flock,
		logw:      logw,
		logSize:   logSize,
		dataPaths: fullDataPaths,
	}
	runtime.SetFinalizer(fs, (*MultiDiskStorage).Close)
	return fs, nil
}

// Lock keep instance of MultiDiskStorage unavailable
func (fs *MultiDiskStorage) Lock() (storage.Locker, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.open < 0 {
		return nil, storage.ErrClosed
	}
	if fs.readOnly {
		return &MultiDiskStorageLock{}, nil
	}
	if fs.slock != nil {
		return nil, storage.ErrLocked
	}
	fs.slock = &MultiDiskStorageLock{fs: fs}
	return fs.slock, nil
}

func itoa(buf []byte, i int, wid int) []byte {
	u := uint(i)
	if u == 0 && wid <= 1 {
		return append(buf, '0')
	}

	// Assemble decimal in reverse order.
	var b [32]byte
	bp := len(b)
	for ; u > 0 || wid > 0; u /= 10 {
		bp--
		wid--
		b[bp] = byte(u%10) + '0'
	}
	return append(buf, b[bp:]...)
}

func (fs *MultiDiskStorage) printDay(t time.Time) {
	if fs.day == t.Day() {
		return
	}
	fs.day = t.Day()
	fs.logw.Write([]byte("=============== " + t.Format("Jan 2, 2006 (MST)") + " ===============\n"))
}

func (fs *MultiDiskStorage) doLog(t time.Time, str string) {
	if fs.logSize > logSizeThreshold {
		// Rotate log file.
		fs.logw.Close()
		fs.logw = nil
		fs.logSize = 0
		rename(filepath.Join(fs.path, "LOG"), filepath.Join(fs.path, "LOG.old"))
	}
	if fs.logw == nil {
		var err error
		fs.logw, err = os.OpenFile(filepath.Join(fs.path, "LOG"), os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return
		}
		// Force printDay on new log file.
		fs.day = 0
	}
	fs.printDay(t)
	hour, min, sec := t.Clock()
	msec := t.Nanosecond() / 1e3
	// time
	fs.buf = itoa(fs.buf[:0], hour, 2)
	fs.buf = append(fs.buf, ':')
	fs.buf = itoa(fs.buf, min, 2)
	fs.buf = append(fs.buf, ':')
	fs.buf = itoa(fs.buf, sec, 2)
	fs.buf = append(fs.buf, '.')
	fs.buf = itoa(fs.buf, msec, 6)
	fs.buf = append(fs.buf, ' ')
	// write
	fs.buf = append(fs.buf, []byte(str)...)
	fs.buf = append(fs.buf, '\n')
	n, _ := fs.logw.Write(fs.buf)
	fs.logSize += int64(n)
}

// Log emits a string log
func (fs *MultiDiskStorage) Log(str string) {
	if !fs.readOnly {
		t := time.Now()
		fs.mu.Lock()
		defer fs.mu.Unlock()
		if fs.open < 0 {
			return
		}
		fs.doLog(t, str)
	}
}

func (fs *MultiDiskStorage) log(str string) {
	if !fs.readOnly {
		fs.doLog(time.Now(), str)
	}
}

func (fs *MultiDiskStorage) setMeta(fd storage.FileDesc) error {
	content := fsGenName(fd) + "\n"
	// Check and backup old CURRENT file.
	currentPath := filepath.Join(fs.path, "CURRENT")
	if _, err := os.Stat(currentPath); err == nil {
		b, err := ioutil.ReadFile(currentPath)
		if err != nil {
			fs.log(fmt.Sprintf("backup CURRENT: %v", err))
			return err
		}
		if string(b) == content {
			// Content not changed, do nothing.
			return nil
		}
		if err := writeFileSynced(currentPath+".bak", b, 0644); err != nil {
			fs.log(fmt.Sprintf("backup CURRENT: %v", err))
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	path := fmt.Sprintf("%s.%d", filepath.Join(fs.path, "CURRENT"), fd.Num)
	if err := writeFileSynced(path, []byte(content), 0644); err != nil {
		fs.log(fmt.Sprintf("create CURRENT.%d: %v", fd.Num, err))
		return err
	}
	// Replace CURRENT file.
	if err := rename(path, currentPath); err != nil {
		fs.log(fmt.Sprintf("rename CURRENT.%d: %v", fd.Num, err))
		return err
	}
	// Sync root directory.
	if err := syncDir(fs.path); err != nil {
		fs.log(fmt.Sprintf("syncDir: %v", err))
		return err
	}
	return nil
}

// SetMeta reset meta info for instance of MultiDiskStorage
func (fs *MultiDiskStorage) SetMeta(fd storage.FileDesc) error {
	if !storage.FileDescOk(fd) {
		return storage.ErrInvalidFile
	}
	if fs.readOnly {
		return errReadOnly
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.open < 0 {
		return storage.ErrClosed
	}
	return fs.setMeta(fd)
}

// GetMeta get meta of instance of MultiDiskStorage
func (fs *MultiDiskStorage) GetMeta() (storage.FileDesc, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.open < 0 {
		return storage.FileDesc{}, storage.ErrClosed
	}
	dir, err := os.Open(fs.path)
	if err != nil {
		return storage.FileDesc{}, err
	}
	names, err := dir.Readdirnames(0)
	// Close the dir first before checking for Readdirnames error.
	if ce := dir.Close(); ce != nil {
		fs.log(fmt.Sprintf("close dir: %v", ce))
	}
	if err != nil {
		return storage.FileDesc{}, err
	}
	// Try this in order:
	// - CURRENT.[0-9]+ ('pending rename' file, descending order)
	// - CURRENT
	// - CURRENT.bak
	//
	// Skip corrupted file or file that point to a missing target file.
	type currentFile struct {
		name string
		fd   storage.FileDesc
	}
	tryCurrent := func(name string) (*currentFile, error) {
		b, err := ioutil.ReadFile(filepath.Join(fs.path, name))
		if err != nil {
			if os.IsNotExist(err) {
				err = os.ErrNotExist
			}
			return nil, err
		}
		var fd storage.FileDesc
		if len(b) < 1 || b[len(b)-1] != '\n' || !fsParseNamePtr(string(b[:len(b)-1]), &fd) {
			fs.log(fmt.Sprintf("%s: corrupted content: %q", name, b))
			err := &storage.ErrCorrupted{
				Err: errors.New("leveldb/storage: corrupted or incomplete CURRENT file"),
			}
			return nil, err
		}
		if _, err := fs.MultiStat(filepath.Join(fs.path, fsGenName(fd))); err != nil {
			if os.IsNotExist(err) {
				fs.log(fmt.Sprintf("%s: missing target file: %s", name, fd))
				err = os.ErrNotExist
			}
			return nil, err
		}
		return &currentFile{name: name, fd: fd}, nil
	}
	tryCurrents := func(names []string) (*currentFile, error) {
		var (
			cur *currentFile
			// Last corruption error.
			lastCerr error
		)
		for _, name := range names {
			var err error
			cur, err = tryCurrent(name)
			if err == nil {
				break
			} else if err == os.ErrNotExist {
				// Fallback to the next file.
			} else if isCorrupted(err) {
				lastCerr = err
				// Fallback to the next file.
			} else {
				// In case the error is due to permission, etc.
				return nil, err
			}
		}
		if cur == nil {
			err := os.ErrNotExist
			if lastCerr != nil {
				err = lastCerr
			}
			return nil, err
		}
		return cur, nil
	}

	// Try 'pending rename' files.
	var nums []int64
	for _, name := range names {
		if strings.HasPrefix(name, "CURRENT.") && name != "CURRENT.bak" {
			i, err := strconv.ParseInt(name[8:], 10, 64)
			if err == nil {
				nums = append(nums, i)
			}
		}
	}
	var (
		pendCur   *currentFile
		pendErr   = os.ErrNotExist
		pendNames []string
	)
	if len(nums) > 0 {
		sort.Sort(sort.Reverse(int64Slice(nums)))
		pendNames = make([]string, len(nums))
		for i, num := range nums {
			pendNames[i] = fmt.Sprintf("CURRENT.%d", num)
		}
		pendCur, pendErr = tryCurrents(pendNames)
		if pendErr != nil && pendErr != os.ErrNotExist && !isCorrupted(pendErr) {
			return storage.FileDesc{}, pendErr
		}
	}

	// Try CURRENT and CURRENT.bak.
	curCur, curErr := tryCurrents([]string{"CURRENT", "CURRENT.bak"})
	if curErr != nil && curErr != os.ErrNotExist && !isCorrupted(curErr) {
		return storage.FileDesc{}, curErr
	}

	// pendCur takes precedence, but guards against obsolete pendCur.
	if pendCur != nil && (curCur == nil || pendCur.fd.Num > curCur.fd.Num) {
		curCur = pendCur
	}

	if curCur != nil {
		// Restore CURRENT file to proper state.
		if !fs.readOnly && (curCur.name != "CURRENT" || len(pendNames) != 0) {
			// Ignore setMeta errors, however don't delete obsolete files if we
			// catch error.
			if err := fs.setMeta(curCur.fd); err == nil {
				// Remove 'pending rename' files.
				for _, name := range pendNames {
					if err := os.Remove(filepath.Join(fs.path, name)); err != nil {
						fs.log(fmt.Sprintf("remove %s: %v", name, err))
					}
				}
			}
		}
		return curCur.fd, nil
	}

	// Nothing found.
	if isCorrupted(pendErr) {
		return storage.FileDesc{}, pendErr
	}
	return storage.FileDesc{}, curErr
}

// List list all ft type files info
func (fs *MultiDiskStorage) List(ft storage.FileType) (fds []storage.FileDesc, err error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.open < 0 {
		return nil, storage.ErrClosed
	}
	dir, err := os.Open(fs.path)
	if err != nil {
		return
	}
	names, err := dir.Readdirnames(0)
	// Close the dir first before checking for Readdirnames error.
	if cerr := dir.Close(); cerr != nil {
		fs.log(fmt.Sprintf("close dir: %v", cerr))
	}
	if err == nil {
		for _, name := range names {
			if fd, ok := fsParseName(name); ok && fd.Type&ft != 0 {
				fds = append(fds, fd)
			}
		}
		otherSSTs, oErr := fs.MultiList(ft)
		if oErr == nil {
			fds = append(fds, otherSSTs...)
		} else {
			return nil, oErr
		}
	}
	return
}

// Open open file with fd file description
func (fs *MultiDiskStorage) Open(fd storage.FileDesc) (storage.Reader, error) {
	if !storage.FileDescOk(fd) {
		return nil, storage.ErrInvalidFile
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.open < 0 {
		return nil, storage.ErrClosed
	}
	of, err := fs.MultiOpenFile(filepath.Join(fs.path, fsGenName(fd)), os.O_RDONLY, 0)
	if err != nil {
		if fsHasOldName(fd) && os.IsNotExist(err) {
			of, err = os.OpenFile(filepath.Join(fs.path, fsGenOldName(fd)), os.O_RDONLY, 0)
			if err == nil {
				goto ok
			}
		}
		return nil, err
	}
ok:
	fs.open++
	return &fileWrap{File: of, fs: fs, fd: fd}, nil
}

// Create create an instance of Writer with fd
func (fs *MultiDiskStorage) Create(fd storage.FileDesc) (storage.Writer, error) {
	if !storage.FileDescOk(fd) {
		return nil, storage.ErrInvalidFile
	}
	if fs.readOnly {
		return nil, errReadOnly
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.open < 0 {
		return nil, storage.ErrClosed
	}
	of, err := fs.MultiOpenFile(filepath.Join(fs.path, fsGenName(fd)), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	fs.open++
	return &fileWrap{File: of, fs: fs, fd: fd}, nil
}

// Remove remove file with file description of fd
func (fs *MultiDiskStorage) Remove(fd storage.FileDesc) error {
	if !storage.FileDescOk(fd) {
		return storage.ErrInvalidFile
	}
	if fs.readOnly {
		return errReadOnly
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.open < 0 {
		return storage.ErrClosed
	}
	err := fs.MultiRemove(filepath.Join(fs.path, fsGenName(fd)))
	if err != nil {
		if fsHasOldName(fd) && os.IsNotExist(err) {
			if e1 := os.Remove(filepath.Join(fs.path, fsGenOldName(fd))); !os.IsNotExist(e1) {
				fs.log(fmt.Sprintf("remove %s: %v (old name)", fd, err))
				err = e1
			}
		} else {
			fs.log(fmt.Sprintf("remove %s: %v", fd, err))
		}
	}
	return err
}

// Rename rename the file description
func (fs *MultiDiskStorage) Rename(oldfd, newfd storage.FileDesc) error {
	if !storage.FileDescOk(oldfd) || !storage.FileDescOk(newfd) {
		return storage.ErrInvalidFile
	}
	if oldfd == newfd {
		return nil
	}
	if fs.readOnly {
		return errReadOnly
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.open < 0 {
		return storage.ErrClosed
	}
	return rename(filepath.Join(fs.path, fsGenName(oldfd)), filepath.Join(fs.path, fsGenName(newfd)))
}

// Close close the instance of MultiDiskStorage
func (fs *MultiDiskStorage) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.open < 0 {
		return storage.ErrClosed
	}
	// Clear the finalizer.
	runtime.SetFinalizer(fs, nil)

	if fs.open > 0 {
		fs.log(fmt.Sprintf("close: warning, %d files still open", fs.open))
	}
	fs.open = -1
	if fs.logw != nil {
		fs.logw.Close()
	}
	return fs.flock.release()
}

type fileWrap struct {
	*os.File
	fs     *MultiDiskStorage
	fd     storage.FileDesc
	closed bool
}

func (fw *fileWrap) Sync() error {
	if err := fw.File.Sync(); err != nil {
		return err
	}
	if fw.fd.Type == storage.TypeManifest {
		// Also sync parent directory if file type is manifest.
		// See: https://code.google.com/p/leveldb/issues/detail?id=190.
		if err := syncDir(fw.fs.path); err != nil {
			fw.fs.log(fmt.Sprintf("syncDir: %v", err))
			return err
		}
	}
	return nil
}

func (fw *fileWrap) Close() error {
	fw.fs.mu.Lock()
	defer fw.fs.mu.Unlock()
	if fw.closed {
		return storage.ErrClosed
	}
	fw.closed = true
	fw.fs.open--
	err := fw.File.Close()
	if err != nil {
		fw.fs.log(fmt.Sprintf("close %s: %v", fw.fd, err))
	}
	return err
}

func fsGenName(fd storage.FileDesc) string {
	switch fd.Type {
	case storage.TypeManifest:
		return fmt.Sprintf("MANIFEST-%06d", fd.Num)
	case storage.TypeJournal:
		return fmt.Sprintf("%06d.log", fd.Num)
	case storage.TypeTable:
		return fmt.Sprintf(sstFormat, fd.Num)
	case storage.TypeTemp:
		return fmt.Sprintf("%06d.tmp", fd.Num)
	default:
		panic("invalid file type")
	}
}

func fsHasOldName(fd storage.FileDesc) bool {
	return fd.Type == storage.TypeTable
}

func fsGenOldName(fd storage.FileDesc) string {
	switch fd.Type {
	case storage.TypeTable:
		return fmt.Sprintf("%06d.sst", fd.Num)
	}
	return fsGenName(fd)
}

func fsParseName(name string) (fd storage.FileDesc, ok bool) {
	var tail string
	_, err := fmt.Sscanf(name, "%d.%s", &fd.Num, &tail)
	if err == nil {
		switch tail {
		case "log":
			fd.Type = storage.TypeJournal
		case "ldb", "sst":
			fd.Type = storage.TypeTable
		case "tmp":
			fd.Type = storage.TypeTemp
		default:
			return
		}
		return fd, true
	}
	n, _ := fmt.Sscanf(name, "MANIFEST-%d%s", &fd.Num, &tail)
	if n == 1 {
		fd.Type = storage.TypeManifest
		return fd, true
	}
	return
}

func fsParseNamePtr(name string, fd *storage.FileDesc) bool {
	_fd, ok := fsParseName(name)
	if fd != nil {
		*fd = _fd
	}
	return ok
}

func (fs *MultiDiskStorage) getRealPath(name string) string {
	var fdNum uint64
	fmt.Sscanf(filepath.Base(name), "%d.ldb", &fdNum)
	N := uint64(len(fs.dataPaths))
	return filepath.Join(fs.dataPaths[fdNum%N], fmt.Sprintf(sstFormat, fdNum))
}

func (fs *MultiDiskStorage) expandDataPaths(name string) []string {
	return append([]string{filepath.Dir(name)}, fs.dataPaths...)
}

// MultiGuessOpenFile guess which disk the file located in and open
func (fs *MultiDiskStorage) MultiGuessOpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	var fdNum uint64
	var openErr error
	var fileHandler *os.File
	fmt.Sscanf(filepath.Base(name), "%d.ldb", &fdNum)
	for _, dp := range fs.expandDataPaths(name) {
		fullName := filepath.Join(dp, fmt.Sprintf(sstFormat, fdNum))
		fs.log(fmt.Sprintf("guess open %s", fullName))
		fileHandler, openErr = os.OpenFile(fullName, flag, perm)
		if openErr == nil {
			return fileHandler, nil
		}
	}
	return nil, openErr
}

// MultiGuessStat guess which disk the file located in and stat
func (fs *MultiDiskStorage) MultiGuessStat(name string) (os.FileInfo, error) {
	var fdNum uint64
	var statErr error
	var fileInfo os.FileInfo
	fmt.Sscanf(filepath.Base(name), "%d.ldb", &fdNum)
	for _, dp := range fs.expandDataPaths(name) {
		fullName := filepath.Join(dp, fmt.Sprintf(sstFormat, fdNum))
		fs.log(fmt.Sprintf("guess stat %s", fullName))
		fileInfo, statErr = os.Stat(fullName)
		if statErr == nil {
			return fileInfo, nil
		}
	}
	return nil, statErr
}

// MultiGuessRemove guess which disk the file located in and remove
func (fs *MultiDiskStorage) MultiGuessRemove(name string) error {
	var fdNum uint64
	var removeErr error
	fmt.Sscanf(filepath.Base(name), "%d.ldb", &fdNum)
	for _, dp := range fs.expandDataPaths(name) {
		fullName := filepath.Join(dp, fmt.Sprintf(sstFormat, fdNum))
		fs.log(fmt.Sprintf("guess remove %s", fullName))
		removeErr = os.Remove(fullName)
		if removeErr == nil {
			return nil
		}
	}
	return removeErr
}

// MultiOpenFile open file in multi disks storage
func (fs *MultiDiskStorage) MultiOpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	if len(fs.dataPaths) > 0 && strings.HasSuffix(name, ".ldb") {
		realName := fs.getRealPath(name)
		fileHandler, openErr := os.OpenFile(realName, flag, perm)
		if openErr == nil {
			return fileHandler, nil
		}
		switch openErr.(type) {
		case *os.PathError:
			if flag == os.O_RDONLY { //ready only
				return fs.MultiGuessOpenFile(name, flag, perm)
			}
			return nil, openErr
		default:
			return nil, openErr
		}
	} else {
		return os.OpenFile(name, flag, perm)
	}
}

// MultiRemove remove file in multi disks storage
func (fs *MultiDiskStorage) MultiRemove(name string) error {
	if len(fs.dataPaths) > 0 && strings.HasSuffix(name, ".ldb") {
		realName := fs.getRealPath(name)
		removeErr := os.Remove(realName)
		if removeErr != nil {
			return fs.MultiGuessRemove(name)
		}
		return nil
	}
	return os.Remove(name)
}

// MultiStat stat file in multi disks storage
func (fs *MultiDiskStorage) MultiStat(name string) (os.FileInfo, error) {
	if len(fs.dataPaths) > 0 && strings.HasSuffix(name, ".ldb") {
		realName := fs.getRealPath(name)
		fileInfo, statErr := os.Stat(realName)
		if statErr != nil {
			return fs.MultiGuessStat(name)
		}
		return fileInfo, nil
	}
	return os.Stat(name)
}

// MultiList list file in multi disks storage
func (fs *MultiDiskStorage) MultiList(ft storage.FileType) (fds []storage.FileDesc, err error) {
	for _, path := range fs.dataPaths {
		dir, err := os.Open(path)
		if err != nil {
			return fds, err
		}
		names, err := dir.Readdirnames(0)
		// Close the dir first before checking for Readdirnames error.
		if cerr := dir.Close(); cerr != nil {
			fs.log(fmt.Sprintf("close dir: %v", cerr))
		}
		if err == nil {
			for _, name := range names {
				if fd, ok := fsParseName(name); ok && fd.Type&ft != 0 {
					fds = append(fds, fd)
				}
			}
		}
	}
	return
}
