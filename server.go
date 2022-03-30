package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	//在线用户列表
	OnLineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

//全参构造器
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnLineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

func (this *Server) ServerListen() {
	for {
		msg := <-this.Message

		//将msg发送给全部的在线User
		this.mapLock.Lock()
		for _, cli := range this.OnLineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

//广播消息
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Name + "]" + msg
	this.Message <- sendMsg
}

func (this *Server) UserOnline(user *User) {
	this.mapLock.Lock()
	this.OnLineMap[user.Name] = user
	this.mapLock.Unlock()

	//广播用户上线消息
	this.BroadCast(user, "已上线")
}

func (this *Server) UserOffline(user *User) {
	this.mapLock.Lock()
	delete(this.OnLineMap, user.Name)
	this.mapLock.Unlock()

	//广播用户下线消息
	this.BroadCast(user, "已下线")
}

func sendMsg(user *User, msg string) {
	user.conn.Write([]byte(msg))
}

func (this *Server) DoMessage(user *User, msg string) {
	if msg == "who" {
		//查看当前在线用户都有哪些
		this.mapLock.Lock()
		for _, u := range this.OnLineMap {
			onlineMsg := "[" + u.Name + "]\n"
			sendMsg(user, onlineMsg)
		}
		this.mapLock.Unlock()
	} else {
		this.BroadCast(user, msg)
	}
}

func (this *Server) HandleConn(conn net.Conn) {
	user := NewUser(conn)
	//用户上线，加入onlinemap
	this.UserOnline(user)

	//接收客户端的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				this.UserOffline(user)
				return
			}
			if err != nil && err != io.EOF {
				log.Fatal(err)
			}
			//提取用户的消息，去除换行符
			msg := string(buf[:n-1])
			this.DoMessage(user, msg)
		}
	}()

	//阻塞当前handler
	select {}
}

//启动服务器
func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		log.Fatal(err)
		return
	}
	//close listen socket
	defer listener.Close()

	//启动监听Message进行消息转发
	go this.ServerListen()

	for {
		//accpet
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
			continue
		}
		//do handler
		go this.HandleConn(conn)
	}

}
