package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/slackhq/nebula"
	"github.com/slackhq/nebula/config"
	"github.com/slackhq/nebula/util"
	"google.golang.org/grpc"
)

// A version string that can be set with
//
//	-ldflags "-X main.Build=SOMEVERSION"
//
// at compile-time.
var (
	Build   string
	mlogger *logrus.Logger
)

type AvoidServer struct {
	nebula.UnimplementedAvoidServer
}

func (s *AvoidServer) GetHostInfo(ctx context.Context, req *nebula.GetHostInfoRequest) (*nebula.GetHostInfoResponse, error) {

	if req == nil {
		errMsg := fmt.Sprintf("GetHostInfo: Invalid Request")
		mlogger.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	return &nebula.GetHostInfoResponse{}, nil
}

func main() {
	serviceFlag := flag.String("service", "", "Control the system service.")
	configPath := flag.String("config", "", "Path to either a file or directory to load configuration from")
	configTest := flag.Bool("test", false, "Test the config and print the end result. Non zero exit indicates a faulty config")
	printVersion := flag.Bool("version", false, "Print version")
	printUsage := flag.Bool("help", false, "Print command line usage")

	flag.Parse()

	if *printVersion {
		fmt.Printf("Version: %s\n", Build)
		os.Exit(0)
	}

	if *printUsage {
		flag.Usage()
		os.Exit(0)
	}

	if *serviceFlag != "" {
		doService(configPath, configTest, Build, serviceFlag)
		os.Exit(1)
	}

	if *configPath == "" {
		fmt.Println("-config flag must be set")
		flag.Usage()
		os.Exit(1)
	}

	mlogger = logrus.New()
	mlogger.Out = os.Stdout

	c := config.NewC(mlogger)
	err := c.Load(*configPath)
	if err != nil {
		fmt.Printf("failed to load config: %s", err)
		os.Exit(1)
	}

	var debug bool
	var port int

	flag.IntVar(&port, "port", 12345, "set the Discoveryd control port")
	flag.BoolVar(&debug, "debug", false, "enable extra debug logging")

	portStr := os.Getenv("AVOIDPORT")
	if portStr != "" {
		portInt, err := strconv.Atoi(portStr)
		if err != nil {
			mlogger.Warningf("Failed to convert AVOIDPORT to int, ignored: %v", portStr)
		} else {
			port = portInt
		}
	}

	debugStr := os.Getenv("DEBUG")
	if debugStr != "" {
		debugInt, err := strconv.ParseBool(debugStr)
		if err != nil {
			mlogger.Warningf("Failed to convert DEBUG to bool, ignored: %v", debugStr)
		} else {
			debug = debugInt
		}
	}

	// daemon mode
	if debug {
		mlogger.SetLevel(logrus.DebugLevel)
	} else {
		mlogger.SetLevel(logrus.InfoLevel)
	}

	ctrl, err := nebula.Main(c, *configTest, Build, mlogger, nil)
	if err != nil {
		util.LogWithContextIfNeeded("Failed to start", err, mlogger)
		os.Exit(1)
	}

	if !*configTest {
		go func() {
			mlogger.Infof("starting lincoln api up on port %d", port)

			lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
			if err != nil {
				mlogger.Fatalf("failed to listen: %v", err)
			}

			grpcServer := grpc.NewServer()
			nebula.RegisterAvoidServer(grpcServer, &AvoidServer{})
			grpcServer.Serve(lis)
		}()

		ctrl.Start()
		ctrl.ShutdownBlock()
	}

	os.Exit(0)
}
