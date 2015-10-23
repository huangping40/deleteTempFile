package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

func main() {

	arg_num := len(os.Args)

	if arg_num != 3 {
		fmt.Printf("Usage is \" files_will_be_deleted_in_directory \"  elapsed_time_in_hours(min 3 hours) \n")
		return
	}

	var path string = os.Args[1]
	checkPath(path)

	h, _ := strconv.ParseInt(os.Args[2], 0, 64)
	if h < 3 {
		h = 3
	}

	var beforeTime time.Duration = -(time.Duration(h)) * time.Hour
	checkTime := time.Now().Add(beforeTime)

	files, err := ioutil.ReadDir(os.Args[1])

	if err != nil {
		fmt.Printf("error, %v", err)
		return
	}

	dtf := DeleteTempFile{Path: path}

	fiChan := make(chan os.FileInfo, 1)

	for i := 0; i < 10; i++ {
		go dtf.Delete(fiChan)
	}

	var willDeleteFileNum int32

	for _, f := range files {
		if f.ModTime().Before(checkTime) {
			fiChan <- f
			willDeleteFileNum++
		}
	}

	var delayCount int
	for delayCount < 1000 {
		delayCount++
		if willDeleteFileNum <= dtf.DeletedFileNum {
			break
		}

		time.Sleep(time.Second)
	}

	dtf.ShowResult()
	fmt.Printf(" \n total file is %v, time is %v, game is over \n", len(files), time.Now())
}

func checkPath(path string) {
	if !strings.HasPrefix(path, "/data/") {
		panic(" path " + path + " is illegal,  only subdirectory in  \"/data\"  is permitted")
	}
}

type DeleteTempFile struct {
	StorageSize, FileNum int64
	DeletedFileNum       int32
	Path                 string
}

func (me *DeleteTempFile) Delete(fiChan chan os.FileInfo) {
	for {
		me.deleteAFile(fiChan)
	}
}
func (me *DeleteTempFile) deleteAFile(fiChan chan os.FileInfo) {
	fi := <-fiChan

	defer func() {
		atomic.AddInt32(&me.DeletedFileNum, 1)
		if r := recover(); r != nil {
			fmt.Printf("\n delete file %s panic, Recovered panic: %s ", fi.Name(), r)
		}
	}()

	me.deleteFileFromOS(fi)
}
func (me *DeleteTempFile) deleteFileFromOS(f os.FileInfo) {
	err := os.Remove(me.Path + "/" + f.Name())

	if err != nil {
		fmt.Println("delete file :%s fails, error info is : %v", f.Name, err)
	} else {
		atomic.AddInt64(&me.StorageSize, f.Size())
		atomic.AddInt64(&me.FileNum, 1)

	}
}
func (me *DeleteTempFile) ShowResult() {

	fmt.Printf("\n   save StorageSize: %v Bytes,  delete file number: %v ", me.StorageSize, me.FileNum)

}
