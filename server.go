package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int
	// 在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	// 消息广播channel
	Message chan string
}

// 创建一个Server
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 广播消息方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

// 监听Message广播消息channel的goroutine
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		// 把message发给所有在线User
		this.mapLock.Lock()
		for _, user := range this.OnlineMap { ////////不广播给自己
			user.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) Handler(conn net.Conn) {
	user := NewUser(conn, this) // 优化

	user.Online()

	// 判断用户是否活跃的channel
	isLive := make(chan bool)

	// 接收客户端发送的信息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("conn Read err:", err)
				return
			}

			// 提取用户消息
			msg := string(buf[:n-1])

			// 将消息广播
			user.DoMessage(msg)

			// 任意操作代表用户活跃
			isLive <- true
		}
	}()

	for {
		select {
		case <-isLive:
			// 当前用户活跃，重置定时器
			// 不做处理，为了激活select，更新定时器

		case <-time.After(time.Second * 100):
			user.SendMsg("long time no operate, you out")

			// 销毁资源
			close(user.C)

			// 退出Handler
			return
		}
	}
}

// 启动Server
func (this *Server) Start() {
	// 启动 socket 监听
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net Listen err:", err)
		return
	}

	// 关闭 socket
	defer listener.Close()

	// 启动监听消息线程
	go this.ListenMessager()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener Accept err:", err)
			continue
		}

		// do handler
		go this.Handler(conn)
	}

}

// 客户端异常退出，服务端报错
// 客户端查询用户异常
