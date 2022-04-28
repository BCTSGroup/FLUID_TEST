package utils

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/op/go-logging"
	"io"
	"log"
	"os"
	"syscall"
	"time"
)

var Log = logging.MustGetLogger("BFCLogger")
var format = logging.MustStringFormatter(
	`%{color}%{time:2006-01-02 15:04:05} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)
var errFile *os.File

func init() {
	writer := configLocalFilesystemLogger("test.log")

	// For demo purposes, create two backend for os.Stderr.
	backend1 := logging.NewLogBackend(writer, "", 0)
	backend2 := logging.NewLogBackend(writer, "", 0)

	// For messages written to backend2 we want to add some additional
	// information to the output, including the used log level and the name of
	// the function.
	backend2Formatter := logging.NewBackendFormatter(backend2, format)

	// Only errors and more severe messages should be sent to backend1
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")

	// Set the backends to be used.
	logging.SetBackend(backend1Leveled, backend2Formatter)
	// Set error log file
	configErrlogFile()
}

//切割日志和清理过期日志
func configLocalFilesystemLogger(filePath string) io.Writer {
	writer, err := rotatelogs.New(
		filePath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(filePath),         // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(time.Hour*24*3),     // 文件最大保存时间
		rotatelogs.WithRotationTime(time.Hour*24), // 日志切割时间间隔
	)
	if err != nil {
		log.Fatal("Init log failed, err:", err)
	}
	return writer
}

func configErrlogFile() {
	// open error file
	file, err := os.OpenFile("err.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	errFile = file
	if err != nil {
		log.Fatal("Init error log failed, err:", err)
	}

	// dup2
	err = syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd()))
	if err != nil {
		log.Fatal("Init error log failed, err:", err)
	}
}
