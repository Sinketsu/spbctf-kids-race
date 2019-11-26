package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net"
)

func main() {
	viper.AutomaticEnv()

	lTCP, err := net.Listen("tcp", viper.GetString("LISTEN"))
	if err != nil {
		logrus.WithError(err).Fatalf("Can't listen TCP")
	}

	go TCPListen(lTCP)
	select {}
}
