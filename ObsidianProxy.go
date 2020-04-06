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
	. "sync"
	_ "unicode"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)
/**
	结构来自于TIS的MCDaemon-go项目
	https://github.com/TISUnion/MCDaemon-go
	感谢项目以及项目作者光叔的帮助 	QWQ
*/
/*基础结构*/
type Server struct{
	name   string         //服务器名称
	Stdout *bufio.Reader  //子进程输出
	Cmd    *exec.Cmd      //子进程实例
	stdin  io.WriteCloser //用于关闭输入管道
	stdout io.ReadCloser  //用于关闭输出管道
	lock   Mutex          //输入管道同步锁
}
/*主函数*/
func main() {
	/*初始化*/
	server := Server{}
	/*检查数据文件是否存在*/
	if server.CheckDataFile("./OPRData") == false {
		/*创建文件*/
		println("[ObsidianProxy]数据文件不存在,正在创建...")
		server.CreateDataFile("./OPRData","data.ini")
	}
	fmt.Println("[ObsidianProxy]Version：1.0.0 Obsidian Build 2020/4/6")
	println("[ObsidianProxy]启动服务器")
	/*启动服务器*/
	server.Init()
}
/*初始化函数*/
func (server *Server) Init(){
	/*服务器名称*/
	server.name = "Minecraft服务器"
	/*进程任务*/
	server.Cmd = exec.Command("java","-jar","fabric-server-launch.jar")
	/*输出管道*/
	stdout, _ := server.Cmd.StdoutPipe()
	/*输入管道*/
	server.stdin, _ = server.Cmd.StdinPipe()
	/*启动进程*/
	_ = server.Cmd.Start()
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
			var playerName = server.GetPlayerName(string(d))
			/*生成指令*/
			var _command = "say " + playerName + "加入了服务器"
			/*执行指令*/
			server.Execute(_command)
			/*进入登录验证环节*/
			go server.PlayerJoinEvent(playerName)
		}
		print(string(d))
	}
	/*模拟CMD暂停*/
	bufio.NewReader(os.Stdin).ReadLine()
}
/**
玩家加入服务器流程
加入服务器 --> 判断是否在白名单内 --> 改变玩家模式
--> 玩家输入密码 --> 服务端检测 --> 恢复模式
*/
/*玩家加入服务器后第一个到达的函数*/
func (server *Server)PlayerJoinEvent(playerName string) {
	server.ChangePlayerMode("spectator", playerName )
	if server.CheckWhiteList(playerName) == false {
		server.KickPlayer(playerName,"您不在白名单内，请联系管理员添加白名单")
	}
	if server.CheckPlayerLoagin(playerName) == false {
		server.KickPlayer(playerName,"您输入的密码不正确，请重试")
	}
	server.SeedColorSubTitle(playerName,"欢迎回来","有问题请联系管理员：Hanbings QQ：3219065882","green","yellow")
}
/*验证白名单*/
func (server *Server)CheckWhiteList(playerName string) bool {
	return true
}
/*验证离线登录*/
func (server *Server)CheckPlayerLoagin(playerName string) bool {
	return true
}
/*检查数据文件是否存在*/
func (server *Server)CheckDataFile(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
/*创建数据文件*/
func (server *Server)CreateDataFile(path string,fileName string){
	_ = os.MkdirAll(path,0777)
	var filePath = path + "/" + fileName
	f,err := os.Create(filePath)
	defer f.Close()
	if err !=nil {
		fmt.Println(err.Error())
	}
}
/*封装发送标题（在屏幕中间显示）*/
func (server *Server)SeedTitle(playerName string,message string)  {
	/*title @a(玩家) title(位置) {"text":"message(文字内容)"}*/
	/*后半部分单句复杂，拆分拼接，注意空格*/
	server.Execute("title" + " " + playerName + " " + "title" + " " + "{\"text\":\""+message+"\"}" )
}
/*封装带颜色的标题（在屏幕中间显示）*/
func (server *Server)SeedColorTitle(playerName string,message string,color string)  {
	/*title @a(玩家) title(位置) {"text":"message(文字内容)","color":"color(颜色)"}*/
	/*后半部分单句复杂，拆分拼接，注意空格*/
	/*"title" + " " + playerName + " " + "title" + "\"{\"text\":\"" + message + "\",\"color\":\"" + color +"\"}\""*/
	server.Execute("title" + " " + playerName + " " + "title" + " " + "{\"text\":\"" + message + "\",\"color\":\"" + color +"\"}")
}
/*封装带子标题的标题（在屏幕中间显示）*/
func (server *Server)SeedSubTitle(playerName string,message string,subMessage string){
	server.Execute("title" + " " + playerName + " " + "subtitle" + " " + "{\"text\":\""+subMessage+"\"}" )
	server.Execute("title" + " " + playerName + " " + "title" + " " + "{\"text\":\""+message+"\"}" )
}
/*封装带颜色和子标题的标题（在屏幕中间显示）*/
func (server *Server)SeedColorSubTitle(playerName string,message string,subMessage string,color string,subColor string){
	server.Execute("title" + " " + playerName + " " + "subtitle" + " " + "{\"text\":\"" + subMessage + "\",\"color\":\"" + subColor +"\"}")
	server.Execute("title" + " " + playerName + " " + "title" + " " + "{\"text\":\"" + message + "\",\"color\":\"" + color +"\"}")
}
/*更改玩家游戏模式*/
func (server *Server)ChangePlayerMode(gamemode string , playerName string){
	server.Execute("gamemode" + " " + gamemode + " " + playerName )
}
/*踢出玩家*/
func (server *Server)KickPlayer(playerName string,message string){
	server.Execute("kick " + playerName + " " +message)
}
/*2020/4/4 执行原版指令*/
func (server *Server)Execute(_command string) {
	/*换行达到回车效果*/
	_command = _command + "\n"
	/*资源锁*/
	server.lock.Lock()
	defer server.lock.Unlock()
	/*向子进程输入*/
	_, err := io.WriteString(server.stdin, _command)
	/*致死量报错*/
	if err != nil {
		fmt.Println("[ObsidianProxy]错误", err)
	}
}
/*获取玩家名函数*/
func (server *Server)GetPlayerName(word string) string {
	/*废话*/
	println("[ObsidianProxy]检测到加入游戏")
	/*截取前玩家名前一段字符*/
	var start = strings.Index(word,"[Server thread/INFO]:")
	/*截取玩家名后一段字符*/
	var end = strings.Index(word,"joined")
	/*切割*/
	var playerName = string([]rune(word)[start + 22:end - 1])
	/*返回玩家名*/
	return playerName
}
