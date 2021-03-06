package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/kildevaeld/valse"
	"github.com/kildevaeld/valse/middlewares/lua"

	"github.com/spf13/pflag"
)

func main() {
	//logrus.SetLevel(logrus.DebugLevel)
	status, err := wrappedMain()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
	os.Exit(status)
}

func wrappedMain() (int, error) {

	address := pflag.StringP("address", "H", ":3000", "address")
	debug := pflag.BoolP("debug", "d", false, "debug")

	pflag.Parse()

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	server := valse.New()

	l := lua.New(server, lua.LuaOptions{
		Path:      ".",
		WorkQueue: 20,
	})

	if err := l.Open(); err != nil {
		return 200, nil
	}

	defer l.Close()

	if err := wait(server, *address); err != nil {
		return -1, err
	}

	return 0, nil
}

func wait(serv *valse.Server, addr string) error {

	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	exit_chan := make(chan error)

	go func() {
		logrus.Printf("Valse started on: '%s'", addr)
		exit_chan <- serv.Listen(addr)
	}()

	go func() {
		signal := <-signal_chan
		logrus.Printf("Signal %s. Existing...", signal)
		exit_chan <- nil //serv.Close()
	}()

	err := <-exit_chan

	signal.Stop(signal_chan)
	close(signal_chan)
	close(exit_chan)

	return err
}
