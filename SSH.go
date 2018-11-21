/*
author：Dill
date：2018-11-20
Used to deploy executable files at remote server
*/

package main

import (
	"flag"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type config struct {
	User     string
	PassWord string
	IpPort   string
	BasePath string
}

var localFile string
var env string
var serverConfig *config

func init() {
	flag.StringVar(&localFile, "f", "", "local file")
	flag.StringVar(&env, "v", "", "deploy env")
	flag.Parse()
	if localFile == "" || env == "" {
		log.Fatal("usage: octopus -f localFile -v gump_dev")
	}

	// TODO:read config file
	serverConfig = new(config)
	switch env {
	case "gump_dev":
		serverConfig.User = "aa"
		serverConfig.PassWord = "bb"
		serverConfig.IpPort = "127.0.0.1:22"
		serverConfig.BasePath = "/aa/server"
	case "ko_dev":
		serverConfig.User = "cc"
		serverConfig.PassWord = "dd"
		serverConfig.IpPort = "127.0.0.2:22"
		serverConfig.BasePath = "/aa/server"
	case "gump_test":
		serverConfig.User = "ee"
		serverConfig.PassWord = "ff"
		serverConfig.IpPort = "127.0.0.3:22"
		serverConfig.BasePath = "/aa/goserver"
	default:
		log.Fatal("parameter v must be one of gump_de, ko_dev and gump_dev")
	}

	// 打印行号
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {

	fmt.Println("**** start deploy ****")
	fmt.Printf("server config: %#v", serverConfig)

	// 获取上传文件绝对路径
	localFilePath, err := filepath.Abs(localFile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("local file path:", localFilePath)

	localFileName := path.Base(localFilePath)
	localFileDir := path.Dir(localFilePath)

	//编译
	fmt.Println(localFileName, "start compiling...")
	cmd := exec.Command("/bin/bash", "-c", "env GOOS=linux GOARCH=amd64 go build -pkgdir "+localFileDir)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	fmt.Println(localFileName, "compile success")

	// 建立 ssh 连接
	client, err := sshNewClient(serverConfig)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	// open an SFTP session over an existing ssh connection
	sftp, err := sftp.NewClient(client)
	if err != nil {
		log.Fatal(err)
	}
	defer sftp.Close()

	fmt.Println("start kill remote process...")
	// 杀死远程进程
	verifyCmd := fmt.Sprintf("ps -ef | grep -w %s | grep -v grep | wc -l", localFileName)
	preNum, err := sshCmdOutput(client, verifyCmd)
	if n, _ := strconv.Atoi(preNum); n > 0 {
		killCmd := fmt.Sprintf("kill $(ps -ef | grep -w %s | grep -v grep | awk '{print $2}')", localFileName)
		if err := sshCmdRun(client, killCmd); err != nil {
			log.Fatal(err)
		}
		// 验证远程进程是否被杀死
		processNum, err := sshCmdOutput(client, verifyCmd)
		if err != nil {
			log.Fatal(err)
		}
		if n, err := strconv.Atoi(processNum); n > 0 || err != nil {
			log.Fatal("kill remote process fail!", err)
		}
	}
	fmt.Println("kill remote process success")

	fmt.Println("start transfer file to remote server...")
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		log.Fatal(err)
	}

	remoteFile := path.Join(serverConfig.BasePath, localFileName, localFileName)
	dstFile, err := sftp.Create(remoteFile)
	if err != nil {
		log.Fatal(err)
	}

	// copy source file to destination file
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		log.Fatal(err)
	}
	fmt.Println("transfer file to remote server success")
	srcFile.Close()
	dstFile.Close()

	// 添加执行权限
	if err := sftp.Chmod(remoteFile, 0755); err != nil {
		log.Fatal(err)
	}

	fmt.Println("start exec remote process...")
	// 执行远端程序
	remoteLog := path.Join(serverConfig.BasePath, localFileName, "nohup.out")
	/*
			如果在脚本中执行 nohup 命令，并且没有进行任何重定向，那么终端上就会弹出
		“nohup: ignoring input and appending output to ’nohup.out’”, 默认 nohup 需要等待一个键入来结束，
		按任意键都行。需要这样一个输入的原因就是没有指定 nohup 接收输出流的位置，
		也就是为什么提示显示将输出放置到目录下的 nohup.out 里。
		要让这个提醒消失也很简单，就是指定输出流向。
		例如：nohup ls / >/dev/null 2>&1 &
	*/
	if err := sshCmdRun(client, "nohup "+remoteFile+" > "+remoteLog+" 2>&1 &"); err != nil {
		log.Fatal(err)
	}
	// 验证程序是否执行
	outNum, err := sshCmdOutput(client, fmt.Sprintf("ps -ef | grep -w %s | grep -v grep | wc -l", localFileName))
	if err != nil {
		log.Fatal(err)
	}
	if n, err := strconv.Atoi(outNum); n != 1 || err != nil {
		log.Fatal("exec remote process fail ", err)
	}
	fmt.Println("exec remote process success")

	fmt.Println("**** deploy success ****")
}

func sshNewClient(con *config) (*ssh.Client, error) {
	// 建立 ssh 连接
	config := &ssh.ClientConfig{
		User: con.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(con.PassWord),
		},

		// 需要验证服务端，不做验证返回nil就可以，点击 HostKeyCallback 看源码就知道了
		//HostKeyCallback: ssh.FixedHostKey(hostkey),
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	client, err := ssh.Dial("tcp", con.IpPort, config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func sshCmdRun(client *ssh.Client, cmd string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	if err := session.Run(cmd); err != nil {
		log.Println(cmd, err)
		return err
	}
	return nil
}

func sshCmdOutput(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	out, err := session.Output(cmd)
	if err != nil {
		log.Println(cmd, err)
		return "", err
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}
