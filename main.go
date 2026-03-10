package main

import (
	"GenshinSymlinker/utils"
	"GenshinSymlinker/workflow"
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	sourceDir := "原神/星铁/绝区零 源文件夹"
	targetDir := "原神/星铁/绝区零 目标文件夹"
	for !utils.PathExists(sourceDir) {
		fmt.Print("请输入游戏「源文件夹」的路径：")
		reader := bufio.NewReader(os.Stdin)
		sourceDir, _ = reader.ReadString('\n')
		sourceDir = strings.Trim(strings.TrimSpace(sourceDir), "\"")
	}
	for !utils.PathExists(targetDir) {
		fmt.Print("请输入游戏「换服包」的路径：")
		reader := bufio.NewReader(os.Stdin)
		targetDir, _ = reader.ReadString('\n')
		targetDir = strings.Trim(strings.TrimSpace(targetDir), "\"")
	}

	decidePre := func(version string) bool {
		var input string
		next := false
		for strings.ToUpper(input) != "Y" && strings.ToUpper(input) != "N" {
			fmt.Println("检测到版本 " + version + " 的预下载包，是否下载下一版本的换服包（y/N）：")
			fmt.Scanln(&input)
		}
		if strings.ToUpper(input) == "Y" {
			next = true
		}
		return next
	}

	decideCN := func() bool {
		var input string
		for strings.ToUpper(input) != "CN" && strings.ToUpper(input) != "EN" {
			fmt.Print("请输入要下载的换服包类型（CN/EN）：")
			fmt.Scanln(&input)
		}
		switch strings.ToUpper(input) {
		case "EN":
			return false
		case "CN":
			return true
		default:
			return true
		}
	}

	err := workflow.Run(workflow.Options{
		SourceDir: sourceDir,
		TargetDir: targetDir,
		DecideCN:  decideCN,
		DecidePre: decidePre,
		OnEvent: func(ev workflow.Event) {
			switch ev.Type {
			case workflow.EventStepStart:
				if ev.Step == workflow.StepVersion {
					gameType := utils.DetectGame(sourceDir)
					switch gameType {
					case "Genshin":
						fmt.Println("正在加载原神版本信息，请稍候…")
					case "StarRail":
						fmt.Println("正在加载星铁版本信息，请稍候…")
					case "ZZZ":
						fmt.Println("正在加载绝区零版本信息，请稍候…")
					}
				}
				if ev.Step == workflow.StepDownload {
					fmt.Println("检测到「换服包」目录内容为空，将下载换服包…")
				}
				if ev.Step == workflow.StepSymlink {
					fmt.Print("正在创建符号链接…")
				}
			case workflow.EventStepEnd:
				if ev.Step == workflow.StepVersion && ev.Message != "" {
					fmt.Println("版本：", ev.Message)
				}
				if ev.Step == workflow.StepDownload {
					fmt.Println("换服包下载完成！")
				}
			case workflow.EventDownload:
				if ev.Index == 1 {
					fmt.Println("正在下载：")
					fmt.Println("文件数：", ev.Total)
				}
				if ev.FileName != "" {
					fmt.Println(ev.FileName)
				}
				if ev.Err != nil {
					fmt.Println(ev.Err)
				}
			case workflow.EventSymlinkDone:
				fmt.Println("完成！")
			}
		},
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("按回车键退出…")
	os.Stdin.Read(make([]byte, 1))
}
