package main
/*空着这里看起来难受 (￣▽￣)"*/
import (
	"bufio"
	"bytes"
	"fmt"
	"gopkg.in/ini.v1"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	. "sync"
	_ "unicode"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	_ "gopkg.in/ini.v1"
)
/**
	结构来自于TIS的MCDaemon-go项目
	https://github.com/TISUnion/MCDaemon-go
	感谢MCDaemon-go项目以及MCDaemon-go项目作者光兄的帮助 	O(∩_∩)O
*/
/*基础结构*/
type Server struct{
	name   string         //服务器名称
	Stdout *bufio.Reader  //子进程输出
	Cmd    *exec.Cmd      //子进程实例
	stdin  io.WriteCloser //用于关闭输入管道
	stdout io.ReadCloser  //用于关闭输出管道
	lock   Mutex          //输入管道同步锁
	version string 		  //Obsidian版本号
	gameVersion string    //服务器游戏版本号
}
/*主函数*/
func main() {
	/*初始化Server*/
	server := Server{}
	/*检查数据文件是否存在*/
	server.CheckData()
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
	/*软件版本号*/
	server.version = "Version：1.0.0 Obsidian Build 2020/4/6"
	/*服务器游戏版本号*/
	server.gameVersion = "Minecraft 1.14.4"
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
			var playerName = server.GetLoginPlayerName(string(d))
			/*生成指令*/
			var _command = "say " + playerName + "加入了服务器"
			/*执行指令*/
			server.Execute(_command)
			/*更改玩家模式并验证白名单*/
			go server.PlayerJoinEvent(playerName)
		}
		/*接受玩家登录指令*/
		if strings.Contains(string(d),"/login"){
			//TODO 登录部分
		}
		/*登录指令*/
		if strings.Contains(string(d),"/l"){
			//TODO 登录指令别名
		}
		/*注册指令*/
		if strings.Contains(string(d),"/register") {
			//TODO 注册部分
		}
		/*注册指令*/
		if strings.Contains(string(d),"/reg"){
			//TODO 注册别名
		}
		/*打印输出*/
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
	/*先更改玩家模式*/
	server.ChangePlayerMode("spectator", playerName )
	/*使用判断玩家是否在白名单内*/
	if server.CheckPlayerOnWhiteList(playerName) == false {
		/*不在就踢出*/
		server.KickPlayer(playerName,"您不在白名单中，请联系管理员添加您的白名单")
		/*不清楚是不是，此处返回不执行接下来的操作*/
		return
	}
	/*成功进入的话能看到登录提醒标题*/
	server.SeedColorSubTitle(playerName,"欢迎，请输入/login 密码进行登录","使用/reg 密码来进行注册，请不要使用您的常用密码","green","yellow")
}
/*校验白名单*/
func (server *Server)CheckPlayerOnWhiteList(playerName string) bool {
	/*上锁，以免多个同时读取/写入*/
	server.lock.Lock()
	/*加载白名单*/
	var whitelist , err = ini.Load("./OPRData/whitelist.ini")
	/*加载失败处理*/
	if err != nil {
		println("[ObsidianProxy][ERRO]读写白名单文件时发生错误...")
		println("[ObsidianProxy][WARN]请尽快检查和备份白名单文件以免再次创建而进行擦写")
		/*一定要解锁*/
		server.lock.Unlock()
		/*返回错误，踢出玩家，以免进入导致崩溃*/
		return false
	}
	/*判断玩家是否在在白名单的节点上*/
	if whitelist.Section("WhiteList").HasKey(playerName) == true {
		/*判断玩家是不是拉进黑名单 即 玩家名 = false*/
		if strings.Contains(whitelist.Section("WhiteList").Key(playerName).Value(),"true") {
			/*输出下*/
			fmt.Println("[ObsidianProxy]玩家 "+ playerName+" 通过了白名单验证")
			/*记得解锁，否则死锁*/
			server.lock.Unlock()
			/*返回通过*/
			return true
		}
	}
	/*解锁解锁*/
	server.lock.Unlock()
	/*若不在白名单节点返回false*/
	return false
}
/*检查数据文件函数*/
func (server *Server)CheckData() {
	/*因为go判断文件存在比较奇葩，先判断是不是存在文件夹，再判断是否存在文件*/
	if server.CheckDataFolder("./OPRData") == false {
		/*创建文件*/
		println("[ObsidianProxy]数据文件夹不存在,正在创建...")
		println("[ObsidianProxy]数据文件data不存在,正在创建...")
		/*创建玩家数据文件*/
		server.CreateDataFile("./OPRData","data.ini")
		println("[ObsidianProxy]数据文件whitelist不存在,正在创建...")
		/*创建白名单配置*/
		server.CreateDataFile("./OPRData","whitelist.ini")
		/*给白名单配置文件加上WhiteList分区*/
		server.CreateINISection("./OPRData/whitelist.ini","WhiteList")
	}
	/*单独判断是否存在文件*/
	if server.CheckDataFile("./OPRData/data.ini") == false {
		println("[ObsidianProxy]数据文件data不存在,正在创建...")
		server.CreateDataFile("./OPRData","data.ini")
	}
	if server.CheckDataFile("./OPRData/whitelist.ini") == false {
		println("[ObsidianProxy]数据文件whitelist不存在,正在创建...")
		server.CreateDataFile("./OPRData","whitelist.ini")
		/*使用这个函数无需上锁，函数内置 O(∩_∩)O*/
		server.CreateINISection("./OPRData/whitelist.ini","WhiteList")
	}
}
/*检查数据文件夹是否存在*/
func (server *Server)CheckDataFolder(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
/*检查数据文件是否存在*/
func (server *Server)CheckDataFile(path string)bool{
	f, err := os.Open(path)
	if err != nil && os.IsNotExist(err) {
		defer f.Close()
		return false
	}
	return true
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
/*执行原版指令*/
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
/*获取正在登录游戏的（非登录系统，原版登录，关闭正版验证相当于直接进入）玩家名函数*/
func (server *Server)GetLoginPlayerName(word string) string {
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
/*在指定的INI文件中产生指定分区*/
func (server *Server)CreateINISection(INIPath string,sectionName string){
	/*锁锁锁 ╰（‵□′）╯*/
	server.lock.Lock()
	var whitelist , err = ini.Load(INIPath)
	if err != nil {
		println("[ObsidianProxy][ERRO]创建一个INI文件的过程中出现了错误：无法加载 " + INIPath + " 文件")
	}
	_, err = whitelist.NewSection(sectionName)
	if err != nil {
		println("[ObsidianProxy][ERRO]创建一个INI文件分区的过程中出现了错误：无法在 " + INIPath + " 文件中创建分区")
	}
	whitelist.SaveTo(INIPath)
	/*解锁解锁 QWQ*/
	server.lock.Unlock()
}
/*打印软件版本号*/
func (server *Server)PrintVersion() {
	fmt.Println("[ObsidianProxy]"+server.GetVersion())
}
/*打印服务器游戏版本号*/
func (server *Server)PrintGameVersion(){
	fmt.Println("[ObsidianProxy]"+server.GetGameVersion())
}
/*获取软件版本号*/
func (server *Server)GetVersion() string {
	return server.version
}
/*获取服务器游戏版本号*/
func (server *Server)GetGameVersion() string {
	return server.gameVersion
}
/*获取服务器名字*/
func (server *Server)GetServerName() string {
	return server.name
}
/*玩家登录函数*/
func (server *Server)PlayerLogin(){
	//TODO 登录部分
}
/*玩家注册函数*/
func (server *Server)PlayerRegister(){
	//TODO 注册部分
}