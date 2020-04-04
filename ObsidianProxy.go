package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	_ "unicode"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)
/**
	结构来自于TIS的MCDaemon-go项目
	https://github.com/TISUnion/MCDaemon-go
	感谢项目以及项目作者光叔的帮助 	QWQ
*/
type Server struct{
	name              string           //服务器名称
	Stdout            *bufio.Reader    //子进程输出
	Cmd               *exec.Cmd        //子进程实例
	stdin             io.WriteCloser   //用于关闭输入管道
	stdout            io.ReadCloser    //用于关闭输出管道
	lock              sync.Mutex       //输入管道同步锁
}
func (server *Server) Init(){
	/*输入执行的命令*/
	server.Cmd = exec.Command("java","-jar","fabric-server-launch.jar")
	/*输出管道*/
	stdout, _ := server.Cmd.StdoutPipe()
	/*输入管道*/
	server.stdin, _ = server.Cmd.StdinPipe()
	/*启动进程*/
	server.Cmd.Start()
	/*读取子进程*/
	reader := bufio.NewReader(stdout)
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		/*转换CMD的编码为GBK*/
		reader := transform.NewReader(
			bytes.NewReader([]byte(line)),
			simplifiedchinese.GBK.NewDecoder(),
		)
		d, _ := ioutil.ReadAll(reader)
		/*将子进程的内容输出*/
		if strings.Contains(string(d),"joined") {
			/*获取玩家名*/
			var playerName = GetPlayerName(string(d))
			/*生成指令*/
			var _command = "say " + playerName + "加入了服务器"
			/*执行指令*/
			server.Execute(_command)
			/*进入登录验证环节*/
			go PlayerJoinEvent(playerName)
		}
		print(string(d))
	}
	/*模拟CMD暂停*/
	bufio.NewReader(os.Stdin).ReadLine()
}
/*主函数*/
func main() {
	/*启动服务器*/
	server := Server{}
	server.Init()
}
/**
	玩家加入服务器流程
	加入服务器 --> 判断是否在白名单内 --> 改变玩家模式
	--> 玩家输入密码 --> 服务端检测 --> 恢复模式
 */
/*玩家加入服务器后第一个到达的函数*/
func PlayerJoinEvent(playerName string) {

}

/*2020/4/4 执行原版指令*/
func (server *Server) Execute(_command string) {
	/*换行达到回车效果*/
	_command = _command + "\n"
	/*资源锁*/
	server.lock.Lock()
	defer server.lock.Unlock()
	/*向子进程输入*/
	_, err := io.WriteString(server.stdin, _command)
	/*致死量报错*/
	if err != nil {
		fmt.Println("错误", err)
	}
}
/*获取玩家名函数*/
func GetPlayerName(word string) string {
	/*废话*/
	print("检测到加入游戏\n")
	/*截取前玩家名前一段字符*/
	var start = strings.Index(word,"[Server thread/INFO]:")
	/*截取玩家名后一段字符*/
	var end = strings.Index(word,"joined")
	/*转换为rune*/
	println(string([]rune(word)))
	/*切割*/
	var playerName = string([]rune(word)[start + 22:end - 1])
	/*返回玩家名*/
	return playerName
}
