package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	// 监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

func (this *User) Online() {
	// 用户上线，加入OnlineMap
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播上线消息
	this.server.BroadCast(this, "online")
}

func (this *User) Offline() {
	// 用户下线，加入OnlineMap
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// 广播下线消息
	this.server.BroadCast(this, "offline")
}

// 给当前User对应的用户发消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询在线用户
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "online...\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" { //改命令格式
		newName := strings.Split(msg, "|")[1]
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("the name is used\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("the name is renew: " + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 获取对方用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("the username is invalid\n")
			return
		}
		// 根据用户名得到对方User对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("the this is not exist\n")
			return
		}
		// 获取消息内容
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("message is null\n")
			return
		}
		remoteUser.SendMsg(this.Name + ":" + content)
	} else {
		this.server.BroadCast(this, msg)
	}
}

// 监听当前user channel消息的goroutine，有消息就发送给对端客户端
func (this *User) ListenMessage() {
	for msg := range this.C {
		_, err := this.conn.Write([]byte(msg + "\n"))
		if err != nil {
			return
		}
	}
	// 不监听后关闭conn，conn在这里关闭最合适
	err := this.conn.Close()
	if err != nil {
		return
	}
}
