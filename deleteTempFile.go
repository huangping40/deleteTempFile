package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type DeleteTempFile struct {
	StorageSize, FileNum int64
	DeletedFileNum       int32
	Path                 string
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf(" Userage is \" -dir=/data/okk  -hour=3 \" ")
		return
	}

	var path string
	flag.StringVar(&path, "dir", "/data/okk", " files_will_be_deleted_in_directory ")

	var h int64
	flag.Int64Var(&h, "hour", 3, " elapsed_time_in_hours(min 3 hours)")
	flag.Parse()

	if h < 3 {
		h = 3
	}

	err := checkPath(path)
	if err != nil {
		fmt.Printf(" path error, %v", err)
		return
	}

	beginReadDirTime := time.Now()
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Printf("read dir error, %v", err)
		return
	}
	endReadDirTime := time.Now()
	fmt.Printf("read dir cost, %v Secs \n", endReadDirTime.Sub(beginReadDirTime).Seconds())

	var wg sync.WaitGroup
	dtf := DeleteTempFile{Path: path}
	fiChan := make(chan os.FileInfo, 100)

	for i := 0; i < 5; i++ {
		go dtf.Delete(fiChan, &wg)
	}

	checkTime := time.Now().Add(-(time.Duration(h)) * time.Hour)
	for _, f := range files {
		if isInvalidFile(f) {
			continue
		}
		if f.ModTime().Before(checkTime) {
			wg.Add(1)
			fiChan <- f
		}
	}
	wg.Wait()
	dtf.ShowResult()
	fmt.Printf(" \n total file is %v, now is %s, game is over \n", len(files), time.Now().Format(time.RFC850))
}

func checkPath(path string) error {
	if !strings.HasPrefix(path, "/data/") {
		return fmt.Errorf(" path  %s is illegal,  only subdirectory in  \"/data\"  is permitted", path)
	}
	return nil
}

func isInvalidFile(f os.FileInfo) bool {
	if f.IsDir() {
		return true
	}
	name := f.Name()
	flen := len(name)
	if flen < 47 {
		return true
	}
	pointLocation := strings.LastIndex(name, ".")
	fmt.Printf(" suffix len:  %v,  %v", pointLocation, flen-pointLocation)
	if pointLocation > 0 && flen-pointLocation < 45 {
		return true
	}
	return false
}

func (dtf *DeleteTempFile) Delete(fiChan chan os.FileInfo, wg *sync.WaitGroup) {
	for {
		err := dtf.deleteAFile(fiChan, wg)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (dtf *DeleteTempFile) deleteAFile(fiChan chan os.FileInfo, wg *sync.WaitGroup) error {
	f := <-fiChan
	defer wg.Done()

	err := os.Remove(filepath.Join(dtf.Path, f.Name()))
	time.Sleep(time.Millisecond * 100)
	if err != nil {
		return fmt.Errorf("\n delete file :%s fails, error info is : %v", f.Name, err)
	} else {
		atomic.AddInt64(&dtf.StorageSize, f.Size())
		if atomic.AddInt64(&dtf.FileNum, 1)%1000 == 0 {
			fmt.Printf("\n   save StorageSize: %v Bytes,  delete file number: %v ", dtf.StorageSize, dtf.FileNum)
		}
	}
	return nil
}

func (dtf *DeleteTempFile) ShowResult() {
	fmt.Printf("\n   save StorageSize: %v Bytes,  delete file number: %v ", dtf.StorageSize, dtf.FileNum)
}
