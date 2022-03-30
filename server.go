package main

import (
	"fmt"
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

func (this *Server) HandleConn(conn net.Conn) {
	user := NewUser(conn)
	//用户上线，加入onlinemap
	this.mapLock.Lock()
	this.OnLineMap[user.Name] = user
	this.mapLock.Unlock()

	//广播用户上线消息
	this.BroadCast(user, "已上线")

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
