package main

import (
	"fmt"
	"github.com/orvice/shadowsocks-go/mu/user"
	ss "github.com/shadowsocks/shadowsocks-go/shadowsocks"
	"os"
	"strconv"
	"time"
)

var configFile string
var config *ss.Config

func boot() {
	var err error

	// log.SetOutput(os.Stdout)

	err = InitMySqlClient()
	if err != nil {
		Log.Error(err)
		os.Exit(0)
	}
	client := user.GetClient()
	users, err := client.GetUsers()
	if err != nil {
		Log.Error(err)
		os.Exit(0)
	}
	Log.Info(len(users))
	bootUsers(users)
	time.Sleep(30 * time.Second)

	go func() {
		for {
			// check users
			users, err = client.GetUsers()
			if err != nil {
				Log.Error(err)
				// os.Exit(0)
			}
			checkUsers(users)
			Log.Info("check finish...")
			time.Sleep(30 * time.Second)
			Log.Info("wake up...")
		}
	}()
	waitSignal()
}

// 第一次启动
func bootUsers(users []user.User) {
	for _, user := range users {
		Log.Info(user.GetUserInfo())
		err := storage.StoreUser(user.GetUserInfo())
		if err != nil {
			Log.Error(err)
		}
		go runWithCustomMethod(user)
	}
}

// check users
func checkUsers(users []user.User) {
	for _, user := range users {
		Log.Debug("check user for ", user.GetPort())

		isExists, err := storage.Exists(user)
		if err != nil {
			Log.Error(err)
			continue
		}
		if !isExists {
			Log.Info("new user to run", user)
			go runWithCustomMethod(user)
			continue
		}
		sUser, err := storage.GetUserInfo(user)
		if err != nil {
			Log.Error(err)
			continue
		}
		if sUser.Passwd != user.GetPasswd() || sUser.Method != user.GetMethod() {
			Log.Info(fmt.Sprintf("user port [%v] passwd or method change ,restart user...", user.GetPort()))
			passwdManager.del(strconv.Itoa(user.GetPort()))
			go runWithCustomMethod(user)
		}
	}
}