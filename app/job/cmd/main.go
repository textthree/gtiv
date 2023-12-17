package main

import (
	"github.com/text3cn/goodle/goodle"
	"github.com/text3cn/goodle/providers/goodlog"
	"gtiv/app/job"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	goodle.Init()
	if err := job.Init(); err != nil {
		panic(err)
	}

	// job
	j := job.New(job.Conf)
	go j.Consume()
	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		goodlog.Pinkf("Job get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			j.Close()
			goodlog.Pink("Job exit")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
