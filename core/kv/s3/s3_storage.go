// Copyright (c) 2020, Junyi Sun <ccnusjy@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package levels3

import (
	"bytes"
	"errors"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
	"sync"
	"syscall"
	"time"
)

var errFileOpen = errors.New("leveldb/storage: file still open")

const CacheSize = 500

type OpenOption struct {
	Bucket        string
	Path          string
	Ak            string
	Sk            string
	Region        string
	Endpoint      string
	LocalCacheDir string
}

type S3StorageLock struct {
	ms *S3Storage
}

func (lock *S3StorageLock) Unlock() {
	log.Println("Unlock")
	ms := lock.ms
	ms.objStore.Remove("LOCK")
	if ms.slock == lock {
		ms.slock = nil
	}
	return
}

// S3Storage is a s3-backed storage.
type S3Storage struct {
	mu       sync.Mutex
	slock    *S3StorageLock
	meta     storage.FileDesc
	objStore *S3Client
	ramFiles *lru.Cache
	opt      OpenOption
}

// NewS3Storage returns a new s3-backed storage implementation.
func NewS3Storage(opt OpenOption) (storage.Storage, error) {
	rand.Seed(int64(time.Now().Nanosecond()))
	s3Client, err := GetS3Client(opt)
	if err != nil {
		return nil, err
	}
	if opt.LocalCacheDir != "" {
		opt.LocalCacheDir = path.Join(opt.LocalCacheDir, opt.Path)
	} else {
		return nil, errors.New("need a local cache dir for s3 storage")
	}
	ramFileCache, _ := lru.New(CacheSize)
	ms := &S3Storage{
		objStore: s3Client,
		ramFiles: ramFileCache,
		opt:      opt,
	}
	return ms, nil
}

func (ms *S3Storage) uploadFiles() error {
	err := os.MkdirAll(ms.opt.LocalCacheDir, 0755)
	if err != nil {
		return err
	}
	logFiles, err := ioutil.ReadDir(ms.opt.LocalCacheDir)
	for _, logF := range logFiles {
		if logF.Name() == "LOCK" {
			continue
		}
		fullName := path.Join(ms.opt.LocalCacheDir, logF.Name())
		content, err := ioutil.ReadFile(fullName)
		if err != nil {
			return err
		}
		log.Println("Upload", fullName)
		fd, _ := fsParseName(logF.Name())
		err = ms.objStore.PutBytes(fd.String(), content)
		if err != nil {
			return err
		}
		os.Remove(fullName)
	}
	return nil
}

func (ms *S3Storage) Lock() (storage.Locker, error) {
	log.Println("Lock")
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if ms.slock != nil {
		return nil, storage.ErrLocked
	}
	locked, _ := ms.objStore.GetBytes("LOCK")
	lockFileName := path.Join(ms.opt.LocalCacheDir, "LOCK")
	localSession, _ := ioutil.ReadFile(lockFileName)
	if string(locked) != string(localSession) && len(locked) > 0 {
		return nil, storage.ErrLocked
	}
	if len(localSession) > 0 {
		var owerPid, rnd int
		fmt.Sscanf(string(localSession), "%d %d", &rnd, &owerPid)
		selfPid := os.Getpid()
		if owerPid != selfPid && syscall.Kill(owerPid, 0) == nil {
			return nil, storage.ErrLocked
		}
	}
	newSession := fmt.Sprintf("%d %d", rand.Int(), os.Getpid())
	err := ms.objStore.PutBytes("LOCK", []byte(newSession))
	if err != nil {
		return nil, storage.ErrLocked
	}
	err = ioutil.WriteFile(lockFileName, []byte(newSession), 0644)
	if err != nil {
		return nil, storage.ErrLocked
	}
	ms.slock = &S3StorageLock{ms: ms}
	err = ms.uploadFiles()
	if err != nil {
		return nil, err
	}
	return ms.slock, nil
}

func (*S3Storage) Log(str string) {}

func (ms *S3Storage) SetMeta(fd storage.FileDesc) error {
	log.Println("SetMeta", fd)
	if !storage.FileDescOk(fd) {
		return storage.ErrInvalidFile
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.meta = fd
	metaStr := fmt.Sprintf("%d %d", fd.Type, int64(fd.Num))
	err := ms.objStore.PutBytes("CURRENT", []byte(metaStr))
	if err != nil {
		return err
	}
	return nil
}

func (ms *S3Storage) GetMeta() (storage.FileDesc, error) {
	log.Println("GetMeta")
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if ms.meta.Zero() {
		metaStr, err := ms.objStore.GetBytes("CURRENT")
		if err != nil {
			return storage.FileDesc{}, os.ErrNotExist
		}
		var fType int
		var fNum int64
		fmt.Sscanf(string(metaStr), "%d %d", &fType, &fNum)
		meta := storage.FileDesc{Type: storage.FileType(fType), Num: fNum}
		log.Println("GetMeta Remote", meta)
		ms.meta = meta
	}
	if ms.meta.Zero() {
		return storage.FileDesc{}, os.ErrNotExist
	}
	log.Println("GetMeta", ms.meta)
	return ms.meta, nil
}

func (ms *S3Storage) List(ft storage.FileType) ([]storage.FileDesc, error) {
	log.Println("List", ft)
	ms.mu.Lock()
	defer ms.mu.Unlock()
	fdsAll, err := ms.objStore.List()
	if err != nil {
		return nil, err
	}
	var fds []storage.FileDesc
	for _, fd := range fdsAll {
		if fd.Type&ft != 0 {
			fds = append(fds, fd)
		}
	}
	return fds, nil
}

func (ms *S3Storage) Open(fd storage.FileDesc) (storage.Reader, error) {
	log.Println("Open", fd)
	if !storage.FileDescOk(fd) {
		return nil, storage.ErrInvalidFile
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()
	if mfileObj, ok := ms.ramFiles.Get(fd.String()); ok {
		mfile := mfileObj.(*memFile)
		mfile.open = true
		return &memReader{Reader: bytes.NewReader(mfile.Bytes()), ms: ms, m: mfile, fd: fd}, nil
	}
	fname := fd.String()
	data, err := ms.objStore.GetBytes(fname)
	if err != nil {
		data, err = ioutil.ReadFile(path.Join(ms.opt.LocalCacheDir, fd.String()))
		if err != nil {
			return nil, os.ErrNotExist
		}
	}
	m := &memFile{Buffer: *bytes.NewBuffer(data), open: true}
	ms.ramFiles.Add(fd.String(), m)
	return &memReader{Reader: bytes.NewReader(data), ms: ms, m: m, fd: fd}, nil
}

func (ms *S3Storage) Create(fd storage.FileDesc) (storage.Writer, error) {
	log.Println("Create", fd)
	if !storage.FileDescOk(fd) {
		return nil, storage.ErrInvalidFile
	}
	m := &memFile{}
	m.open = true
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.ramFiles.Add(fd.String(), m)
	var cacheFile *os.File
	if (fd.Type == storage.TypeJournal || fd.Type == storage.TypeManifest) && ms.opt.LocalCacheDir != "" {
		var err error
		cacheFile, err = os.Create(path.Join(ms.opt.LocalCacheDir, fd.String()))
		if err != nil {
			return nil, err
		}
	}
	return &memWriter{memFile: m, ms: ms, fd: fd, cacheFile: cacheFile}, nil
}

func (ms *S3Storage) Remove(fd storage.FileDesc) error {
	log.Println("Remove", fd)
	if !storage.FileDescOk(fd) {
		return storage.ErrInvalidFile
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.ramFiles.Remove(fd.String())
	os.Remove(path.Join(ms.opt.LocalCacheDir, fd.String()))
	err := ms.objStore.Remove(fd.String())
	if err != nil {
		return err
	}
	return nil
}

func (ms *S3Storage) Rename(oldfd, newfd storage.FileDesc) error {
	log.Println("Rename", oldfd, newfd)
	if !storage.FileDescOk(oldfd) || !storage.FileDescOk(newfd) {
		return storage.ErrInvalidFile
	}
	if oldfd == newfd {
		return nil
	}
	return nil
}

func (ms *S3Storage) Close() error {
	log.Println("storage Close")
	return nil
}

type memFile struct {
	bytes.Buffer
	open bool
}

type memReader struct {
	*bytes.Reader
	ms     *S3Storage
	m      *memFile
	fd     storage.FileDesc
	closed bool
}

func (mr *memReader) Close() error {
	log.Println("reader Close", mr.fd)
	mr.ms.mu.Lock()
	defer mr.ms.mu.Unlock()
	if mr.closed {
		return storage.ErrClosed
	}
	mr.m.open = false
	mr.ms.ramFiles.Remove(mr.fd.String())
	return nil
}

type memWriter struct {
	*memFile
	ms        *S3Storage
	fd        storage.FileDesc
	cacheFile *os.File
}

func (mw *memWriter) Sync() error {
	log.Println("writer Sync", mw.fd, "len", mw.memFile.Len())
	if mw.fd.Type == storage.TypeManifest && mw.cacheFile != nil {
		return mw.cacheFile.Sync()
	}
	return nil
}

func (mw *memWriter) Write(p []byte) (n int, err error) {
	n, err = mw.memFile.Write(p)
	if mw.cacheFile != nil {
		_, err2 := mw.cacheFile.Write(p)
		if err2 != nil {
			return 0, err2
		}
	}
	return n, err
}

func (mw *memWriter) Close() error {
	curLen := mw.memFile.Len()
	log.Println("writer Close", mw.fd, "len", curLen)
	if mw.cacheFile != nil {
		tErr := mw.cacheFile.Close()
		if tErr != nil {
			return tErr
		}
	}
	fname := mw.fd.String()
	err := mw.ms.objStore.PutBytes(fname, mw.memFile.Bytes())
	if err != nil {
		return err
	}
	mw.memFile.open = false
	if mw.cacheFile != nil {
		os.Remove(mw.cacheFile.Name())
	}
	return nil
}
