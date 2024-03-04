package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asvsfs/go-boilerplate/cmd"
	"go.uber.org/zap"
)

var zlog *zap.Logger

func init() {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	zlog, _ = config.Build()
}

func SetLogger(logger *zap.Logger) {
	zlog = logger
}

func main() {
	rand.Seed(time.Now().UnixNano()) // used in rand.Shuffle since it is faster

	var state byte
	const (
		reconfigure byte = iota
		waitForSignal
	)
	signalCh := make(chan os.Signal, 1)
	for {
		switch state {
		case reconfigure:
			state = waitForSignal

			go func() {
				cmd.Execute()
				signalCh <- syscall.SIGQUIT
			}()

		case waitForSignal:
			signal.Notify(signalCh,
				syscall.SIGHUP,  // reconfigure
				syscall.SIGINT,  // Ctrl+C
				syscall.SIGTERM, // Kubernetes best practices: https://cloud.google.com/blog/products/containers-kubernetes/kubernetes-best-practices-terminating-with-grace
				syscall.SIGQUIT)

			// The docker stop command attempts to stop a running container first by sending a SIGTERM signal to the root process (PID 1) in the container. If the process hasn't exited within the timeout period a SIGKILL signal will be sent.
			// Whereas a process can choose to ignore a SIGTERM, a SIGKILL goes straight to the kernel which will terminate the process. The process never even gets to see the signal.
			// When using docker stop the only thing you can control is the number of seconds that the Docker daemon will wait before sending the SIGKILL:
			// docker stop ----time=30 foo

			// docker kill ----signal=SIGINT foo
			// By default, the docker kill command doesn't give the container process an opportunity to exit gracefully -- it simply issues a SIGKILL to terminate the container. However, it does accept a --signal flag which will let you send something other than a SIGKILL to the container process.
			// For example, if you wanted to send a SIGINT (the equivalent of a Ctrl-C on the terminal) to the container "foo" you could use the following:
			// Unlike the docker stop command, kill doesn't have any sort of timeout period. It issues just a single signal (either the default SIGKILL or whatever you specify with the --signal flag).

			// docker rm ----force foo
			// If your goal is to erase all traces of a running container, then docker rm -f is the quickest way to achieve that. However, if you want to allow the container to shutdown gracefully you should avoid this option.

			sig := <-signalCh
			log.Println("signal received:", sig)
			switch sig {
			case syscall.SIGHUP: // reconfigure logic here
				state = reconfigure

			case syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM:
				// clean up then exit, terminating with grace

				return
			}
		}
	}
}
