package main

import (
	"bufio"
	"fmt"
	"github.com/go-redis/redis/v7"
	uuid2 "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"math/rand"
	"net"
	"strings"
	"time"
)

func TCPListen(l net.Listener) {
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			logrus.WithError(err).Errorf("Can't accept TCP connection")
		}

		go TCPHandler(conn)
	}
}

func TCPHandler(conn net.Conn) {
	defer conn.Close()
	rand.Seed(time.Now().UnixNano())

	client := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})
	uuid := uuid2.NewV4().String()
	defer func() {
		count := client.SCard(uuid).Val()
		client.SPopN(uuid, count)
	}()

	fmt.Fprintln(conn, "RUCTFE* Check system! Pass flags one per line:")

	var reader = bufio.NewReader(conn)
	for {
		flag, err := reader.ReadString('\n')
		if err != nil {
			logrus.WithError(err).Errorf("Can't read string")
			break
		}
		flag = strings.TrimSpace(flag)

		if len(flag) == 0 {
			continue
		}

		go FlagAdd(conn, flag, client, uuid)
	}
}

func FlagAdd(conn net.Conn, newFlag string, client *redis.Client, uuid string) {
	flags := client.SMembers(uuid).Val()
	for _, flag := range flags {
		if flag == newFlag {
			fmt.Fprintf(conn, "Flag already sent.")
			return
		}
	}

	result := client.SAdd(uuid, newFlag)
	if result.Val() == 1 {
		fmt.Fprintf(conn, "Flag accepted. %v points earned\n", rand.Float64() * 20)
	} else {
		fmt.Fprintf(conn, "Wow, race condition detected! Your flag is %v\n", viper.GetString("FLAG"))
	}
}
