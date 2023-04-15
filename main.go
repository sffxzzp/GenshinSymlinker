package main

import (
	"GenshinSymlinker/genshin"
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func pathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("%s 未找到！\n", path)
		return false
	}
	return true
}

func isDirEmpty(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return true
	}
	defer f.Close()
	_, err = f.Readdir(1)
	if err == nil {
		return false
	}
	if err == io.EOF {
		return true
	}
	return true
}

func main() {
	sourceDir := "Genshin Impact game"
	targetDir := "YuanShen"
	for !pathExists(sourceDir) {
		fmt.Print("请输入「源文件夹」的路径（支持拖放）：")
		reader := bufio.NewReader(os.Stdin)
		sourceDir, _ = reader.ReadString('\n')
		sourceDir = strings.Trim(strings.TrimSpace(sourceDir), "\"")
	}
	for !pathExists(targetDir) {
		fmt.Print("请输入「换服包」的路径（支持拖放）：")
		reader := bufio.NewReader(os.Stdin)
		targetDir, _ = reader.ReadString('\n')
		targetDir = strings.Trim(strings.TrimSpace(targetDir), "\"")
	}
	if isDirEmpty(targetDir) {
		fmt.Println("检测到「换服包」目录内容为空，将下载换服包…")
		fmt.Println("正在加载原神版本信息，请稍候…")
		ys := genshin.New()
		ys.Compare(true)
		fmt.Println("版本：", ys.Version)
		var input string
		for strings.ToUpper(input) != "CN" && strings.ToUpper(input) != "EN" {
			fmt.Print("请输入要下载的换服包类型（CN/EN）：")
			fmt.Scanln(&input)
		}
		var isCN bool
		switch strings.ToUpper(input) {
		case "EN":
			isCN = false
		case "CN":
			isCN = true
		default:
			isCN = true
		}
		ys.Download(isCN, targetDir)
		fmt.Println("换服包下载完成！")
	}
	fmt.Print("正在创建符号链接…")
	d := genshin.NewDiff()
	d.Init(sourceDir, targetDir)
	d.CreateSymlinks()
	fmt.Println("完成！")
	fmt.Println("按回车键退出…")
	os.Stdin.Read(make([]byte, 1))
}
