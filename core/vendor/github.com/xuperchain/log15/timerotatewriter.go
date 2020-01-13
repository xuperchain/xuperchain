/**
* Time based log rotation Writer.
* Could rotate log file every `rotateInterval` minutes.
**/

package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sort"
	"time"
	"path/filepath"
	"regexp"
)

type TimeRotateWriter struct{
	filename        string
	maxBackups      int     // max log files
	rotateInterval  int     // in minutes

	file            io.WriteCloser
	mutex           sync.Mutex
	rotateAt        int64
	intervalInSeconds   int64
}

func NewTimeRotateWriter(filename string, interval int, backupCount int) (*TimeRotateWriter, error) {
	fullname,err := filepath.Abs(filename)
	if err != nil{
		return nil, err
	}

	wr := TimeRotateWriter{
		filename:       fullname,
		maxBackups:     backupCount,
		rotateInterval: interval,
	}

	// init rotate time
	wr.intervalInSeconds = int64(interval * 60)
	wr.calcNextRotateTime()

	// open file to write
	err = wr.openFile();
	return &wr, err
}

// implements Write interface of io.Writer
func (wr *TimeRotateWriter) Write(data []byte) (succBytes int, err error){
	wr.mutex.Lock()
	defer wr.mutex.Unlock()

	// Open log file
	if err := wr.openFile(); err != nil{
		return 0, err
	}

	// check if the right time to rotate log
	if wr.shouldRotate(){
		if err := wr.rotate(); err != nil{
			return 0, err
		}
	}

	// write log
	return wr.file.Write(data)
}

// Close of WriterCloser
func (wr *TimeRotateWriter) Close() (err error) {
	if err = wr.file.Close(); err != nil {
		return
	}
	wr.file = nil
	return
}

func (wr *TimeRotateWriter) openFile() error{
	if wr.file == nil{
		// mkdir if not exist
		path, _ := filepath.Split(wr.filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.Mkdir(path, 0755)
			os.Chmod(path, 0755)
		}

		// open file for append and write
		fd, err := os.OpenFile(wr.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil{
			return err
		}
		wr.file = fd
	}
	return nil
}

// check if should rotate the log
func (wr *TimeRotateWriter) shouldRotate() bool {
	return time.Now().Unix() >= wr.rotateAt
}

// calculate the next log rotation time
func (wr *TimeRotateWriter) calcNextRotateTime() {
	currentTime := time.Now().Unix()

	timestruct := time.Unix(currentTime, 0)
	//currentHour := timestruct.Hour()
	currentMinute := timestruct.Minute()
	currentSecond := timestruct.Second()

	// elegent split time for hourly rotation
	if wr.rotateInterval % 60 == 0 {
		wr.rotateAt = int64(currentTime + wr.intervalInSeconds - 
			(int64(currentSecond) + int64(currentMinute) * 60))
	} else {
		wr.rotateAt = int64(currentTime - int64(currentSecond) + wr.intervalInSeconds)
	}
}

// do log rotation
func (wr *TimeRotateWriter) rotate() (err error) {
	if err = wr.Close(); err != nil{
		return err
	}

	dstTime := wr.rotateAt - wr.intervalInSeconds
	dstPath := wr.filename + "." + time.Unix(dstTime, 0).Format("200601021504")

	if _, err := os.Stat(dstPath); err == nil {
		os.Remove(dstPath)
	}

	if err = os.Rename(wr.filename, dstPath); err != nil{
		return err
	}

	if wr.maxBackups > 0 {
		wr.deleteExpiredFiles()
	}

	wr.calcNextRotateTime()
	
	err = wr.openFile()
	return err
}

// delete expired log files
func (wr *TimeRotateWriter) deleteExpiredFiles() {
	allfiles := make([]string, 0, 50)
	path, fname := filepath.Split(wr.filename)

	// compile log file regex
	regstr := fname + ".\\d*$"
	fileRegex, err := regexp.Compile(regstr)
	if err != nil{
		fmt.Println("regstr compile failed, regstr=", regstr)
		return
	}

	// iterate all files in the directory
	err = filepath.Walk(path, func(curpath string, info os.FileInfo, err error) error {
		if err != nil{
			fmt.Println("walk error! err=", err)
			return nil
		}

		if info.IsDir(){
			return nil
		}

		if matched := fileRegex.MatchString(info.Name()); matched {
			allfiles = append(allfiles, curpath)
		}

		return nil
	})

	if err != nil {
		return
	}

	// sort files by name
	sort.Strings(allfiles)

	// remove expired files
	fileCount := len(allfiles)
	if fileCount > wr.maxBackups{
		for i := 0; i <  fileCount - wr.maxBackups; i++ {
			err := os.Remove(allfiles[i])
			if err != nil {
				fmt.Println("remove file failed!, file=", allfiles[i])
			}
		}
	}
}