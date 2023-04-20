package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"syscall"
)

const CurrentVersion = "v0.0.1"

var (
	configFileFlag = flag.String(
		"config",
		"",
		"config file (yaml format)",
	)
	logLevelFlag = flag.String(
		"loglevel",
		"info",
		"Set log level.",
	)
	versionFlag = flag.Bool(
		"version",
		false,
		"Print version info.",
	)
)

var log = &logger.Logger{
	Out: os.Stdout,
	Formatter: &logger.TextFormatter{
		CallerPrettyfier: func(caller *runtime.Frame) (function string, file string) {
			function = ""
			_, filename_ := path.Split(caller.File)
			file = fmt.Sprintf("%s:%d", filename_, caller.Line)
			return
		},
		TimestampFormat: "2006-01-02T15:04:05",
	},
	Level:        logger.DebugLevel,
	ReportCaller: true,
}

func printVersion() {
	fmt.Println(CurrentVersion)
}

func main() {
	// Exit on some signals.
	termSig_ := make(chan os.Signal)
	signal.Notify(termSig_, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig_ := <-termSig_
		fmt.Printf("*** Terminating from signal [%+v] ***\n", sig_)
		os.Exit(0)
	}()
	flag.Usage = func() {
		_, execPath_ := filepath.Split(os.Args[0])
		_, _ = fmt.Fprint(os.Stderr, "Simple HTTP (TLS Terminator) Reverse Proxy.\n\n")
		_, _ = fmt.Fprint(os.Stderr, "Version: "+CurrentVersion+".\n\n")
		_, _ = fmt.Fprintf(os.Stderr, "Usage:\n\n  %s [options]\n\nOptions:\n\n", execPath_)
		flag.PrintDefaults()
	}
	flag.Parse()
	if *configFileFlag != "" && PathExists(*configFileFlag) {
		ReadConfigFromFile(*configFileFlag)
	} else {
		flag.Usage()
		os.Exit(1)
	}
	if *versionFlag {
		printVersion()
		return
	}
	fmt.Println("*** Starting ***")

	// Set the loglevel
	logLevel_, err := logger.ParseLevel(*logLevelFlag)
	if err != nil {
		log.Warnf("invalid log level: %v", err)
	}
	log.SetLevel(logLevel_)

	// start server
	serve()

	os.Exit(0)
}

func serve() {
	if logLevel_, err := logger.ParseLevel(*logLevelFlag); err == nil && logLevel_ >= logger.DebugLevel {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	var waitChs_ []chan error

	for _, svcConf := range ExecConfig {
		log.Infof("Starting service [%#v]", svcConf)
		waitCh_ := make(chan error)
		conf := svcConf
		go startServer(&conf, waitCh_)
		waitChs_ = append(waitChs_, waitCh_)
	}

	for _, waitCh := range waitChs_ {
		err := <-waitCh
		if err != nil {
			log.Errorf("Error: %v", err)
		}
	}
}

func startServer(conf *ServiceConfigModel, waitCh chan error) {
	router_ := gin.Default()
	err := router_.SetTrustedProxies([]string{"0.0.0.0/0", "::/0"})
	if err != nil {
		waitCh <- err
		return
	}
	router_.RemoteIPHeaders = []string{"X-Real-IP"}
	router_.UseH2C = true

	handler_ := NewGinHandler(conf)

	router_.GET("/*path", handler_.HandleFun)
	router_.OPTIONS("/*path", handler_.HandleFun)
	router_.HEAD("/*path", handler_.HandleFun)

	router_.POST("/*path", handler_.HandleFun)
	router_.PUT("/*path", handler_.HandleFun)
	router_.DELETE("/*path", handler_.HandleFun)
	router_.PATCH("/*path", handler_.HandleFun)

	err = router_.Run(conf.Listen)
	waitCh <- err
}
