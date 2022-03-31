package main

import (
	"flag"
	"fmt"
	"io"
	"log"
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
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		log.Fatal(err)
		return nil
	}

	client.conn = conn

	return client
}

var serverIp string
var serverPort int

func init() {
	//可以添加命令行参数 -ip 指定IP地址
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址（默认是127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8000, "设置服务器端口号（默认是8000）")
}

func (client *Client) menu() bool {
	var flag int

	fmt.Println("0. 退出")
	fmt.Println("1. 公聊模式")
	fmt.Println("2. 私聊模式")
	fmt.Println("3. 更新用户名")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>请输入合法范围内的数字<<<")
		return false
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println(">>>>请输入用户名：")
	fmt.Scanln(&client.Name)
	msg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(msg))
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

func (client *Client) PublicChat() {
	var msg string
	fmt.Println(">>>请输入聊天内容（exit退出）：")
	fmt.Scanln(&msg)
	for msg != "exit" {
		if len(msg) != 0 {
			_, err := client.conn.Write([]byte(msg + "\n"))
			if err != nil {
				log.Fatal(err)
				break
			}
		}
		msg = ""
		fmt.Println(">>>请输入聊天内容（exit退出）：")
		fmt.Scanln(&msg)
	}
}

func (client *Client) SelectUsers() {
	msg := "who\n"
	_, err := client.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("SelectUsers Error")
	}
}

func (client *Client) PrivateChat() {
	var userName string
	var msg string

	client.SelectUsers()
	fmt.Println(">>>>>请输入聊天对象[用户名]，exit退出: ")
	fmt.Scanln(&userName)

	for userName != "exit" {
		fmt.Println(">>>请输入消息内容，exit退出: ")
		fmt.Scanln(&msg)
		for msg != "exit" {
			if len(msg) != 0 {
				m := "to|" + userName + "|" + msg + "\n\n"
				_, err := client.conn.Write([]byte(m))
				if err != nil {
					fmt.Println("conn write error:", err)
					break
				}
			}
			msg = ""
			fmt.Println(">>>请输入消息内容，exit退出: ")
			fmt.Scanln(&msg)
		}
		client.SelectUsers()
		fmt.Println(">>>>>请输入聊天对象[用户名]，exit退出: ")
		fmt.Scanln(&userName)
	}
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}

		switch client.flag {
		case 1:
			client.PublicChat()
		case 2:
			client.PrivateChat()
		case 3:
			client.UpdateName()
		}
	}
}

func (client *Client) DealResponse() {
	//io.Copy()会永久监听
	if _, err := io.Copy(os.Stdout, client.conn); err != nil {
		log.Fatal(err)
	}
}

//同一个包下可以有多个main函数，只不过编译的时候不能一起编译
func main() {
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("链接服务器失败")
		return
	}
	fmt.Println("链接服务器成功")
	go client.DealResponse()
	client.Run()
}
