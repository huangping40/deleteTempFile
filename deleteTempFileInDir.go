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

var (
	path = flag.String("dir", "", "  制定目录，这个目录下的文件将被删除")
	hour = flag.Int64("hour", 3, "  时间间隔（小时），只有比这个时间更早的文件才被删除,文件名词要求 47个字符以上，且 文件后缀长度大于等于40")
)

func checkFlags() {
	if !strings.HasPrefix(*path, "/data") {
		fmt.Printf(" illeage dir : %v \n ", *path)
		os.Exit(2)
	}

	if *hour < 3 {
		*hour = 0
	}
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("cat /data/test/readme  |  xargs -i -t  ./deleteTempFileInFiles  -dir=/data/test -filename={}  -hour=2")
	fmt.Println()
	flag.PrintDefaults()
	os.Exit(0)
}
func main() {
	flag.Usage = usage
	flag.Parse()
	checkFlags()

	beginReadDirTime := time.Now()
	files, err := ioutil.ReadDir(*path)
	if err != nil {
		fmt.Printf("read dir error, %v", err)
		return
	}
	endReadDirTime := time.Now()
	fmt.Printf("read dir cost, %v Secs \n", endReadDirTime.Sub(beginReadDirTime).Seconds())

	var wg sync.WaitGroup
	dtf := DeleteTempFile{Path: *path}
	fiChan := make(chan os.FileInfo, 100)

	for i := 0; i < 3; i++ {
		go dtf.Delete(fiChan, &wg)
	}

	checkTime := time.Now().Add(-(time.Duration(*hour)) * time.Hour)
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
	time.Sleep(time.Millisecond * 200)
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
