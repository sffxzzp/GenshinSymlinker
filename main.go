package main

import (
	"GenshinSymlinker/genshin"
	"bufio"
	"fmt"
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
	d := genshin.NewDiff()
	d.Init(sourceDir, targetDir)
	d.CreateSymlinks()
}
