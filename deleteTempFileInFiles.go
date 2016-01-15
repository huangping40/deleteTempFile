package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DeleteTempFile struct {
	StorageSize, FileNum int64
	DeletedFileNum       int32
	Path                 string
}

var (
	filename = flag.String("filename", "", "文件名词要求 47个字符以上，且 文件后缀长度大于等于40")
	dir2     = flag.String("dir", "", "  制定目录，这个目录下的文件将被删除")
	hour     = flag.Int64("hour", 3, "  时间间隔（小时），只有比这个时间更早的文件才被删除")
)

func checkFlags() {
	if !strings.HasPrefix(*dir2, "/data") {
		fmt.Printf(" illeage dir : %v \n ", *dir2)
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

	if isInvalidFile2(*filename) {
		return
	}

	f, err := os.Open(filepath.Join(*dir2, *filename))
	if err != nil {
		return
	}

	fi, err := f.Stat()
	f.Close()

	if err != nil {
		return
	}
	if fi.IsDir() {
		return
	}

	checkTime := time.Now().Add(-(time.Duration(*hour)) * time.Hour)
	if fi.ModTime().Before(checkTime) {
		if err = os.Remove(filepath.Join(*dir2, *filename)); err != nil {
			fmt.Printf(" remove  file: %v failed, %v", *filename, err)
		}
	}
	time.Sleep(time.Millisecond * 200)
}

func isInvalidFile2(filename string) bool {

	flen := len(filename)
	if flen < 47 {
		return true
	}
	pointLocation := strings.LastIndex(filename, ".")

	if pointLocation < 0 {
		return true
	}

	if flen-pointLocation < 40 {
		return true
	}
	return false
}
