package main

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	_ "unicode"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func main() {
	// 输入执行的命令
	cmd := exec.Command("java","-jar","fabric-server-launch.jar")

	// 获取子进程标准输出
	stdout, _ := cmd.StdoutPipe()
	stdin, _ := cmd.StdinPipe()
	// 执行命令
	cmd.Start()

	// 读取子进程
	reader := bufio.NewReader(stdout)
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		// 转换CMD的编码为GBK
		reader := transform.NewReader(
			bytes.NewReader([]byte(line)),
			simplifiedchinese.GBK.NewDecoder(),
		)
		d, _ := ioutil.ReadAll(reader)
		// 将子进程的内容输出
		var s = string(d)
		if strings.Contains(s,"joined") {
			var _command = "kick @a 岁月静好，只是有人在替我们负重前行，为英雄默哀"
			_command = _command + "\n"
			_, err2 = io.WriteString(stdin, _command)
		}
		print(s)
	}
	// 模拟CMD暂停
	bufio.NewReader(os.Stdin).ReadLine()
}
