package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	Client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net Dial err:", err)
		return nil
	}

	Client.conn = conn

	return Client
}

func (client *Client) menu() bool {
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("choose is invalid")
		return false
	}

}

func (client *Client) PublicChat() {
	var chatMsg string

	fmt.Println(">>>>>input msg | exit")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>>input msg | exit")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) SelectUser() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("err:", err)
		return
	}
}

func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SelectUser()
	fmt.Println(">>>>> select user | exit")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>> input msg | exit")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("err:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>>>>input msg | exit")
			fmt.Scanln(&chatMsg)
		}

		client.SelectUser()
		fmt.Println(">>>>> select user | exit")
		fmt.Scanln(&remoteName)
	}
}

func (client *Client) UpdateName() {
	fmt.Println(">>>>> input the name:")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("update err:", err)
	}
	fmt.Println("update success")
}

func (client *Client) DealResponse() {
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
			// 菜单选择无效，阻塞
		}

		switch client.flag {
		case 1:
			fmt.Println("公聊模式")
			client.PublicChat()
		case 2:
			fmt.Println("私聊模式")
			client.PrivateChat()
		case 3:
			fmt.Println("更新用户名")
			client.UpdateName()
		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "set server ip")
	flag.IntVar(&serverPort, "port", 8888, "set server port")
}

func main() {
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>> connect fail")
		return
	}

	go client.DealResponse() //////////////

	fmt.Println(">>>>> connect success")

	client.Run()
}
