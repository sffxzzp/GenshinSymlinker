package main

import (
	"fmt"
	"strings"

	"GenshinSymlinker/utils"
)

func main() {
	var gs *utils.Genshin
	var sr *utils.StarRail
	downTypeUI := map[int]string{1: "原神", 2: "星铁"}
	var typeSelect int
	for typeSelect != 1 && typeSelect != 2 {
		fmt.Print("游戏\n1. " + downTypeUI[1] + "\n2. " + downTypeUI[2] + "\n\n请选择（1/2）：")
		fmt.Scanln(&typeSelect)
	}
	if typeSelect == 1 {
		fmt.Println("正在加载" + downTypeUI[typeSelect] + "版本信息，请稍候…")
		gs = utils.NewGenshin()
		gs.Compare(false)
	} else {
		fmt.Println("正在加载" + downTypeUI[typeSelect] + "版本信息，请稍候…")
		sr = utils.NewStarRail()
		sr.Compare(false)
	}
	var input string
	for input != "CN" && input != "EN" {
		fmt.Print("要下载的换服包类型（CN/EN）：")
		fmt.Scanln(&input)
		input = strings.ToUpper(input)
	}
	var isCN bool
	switch input {
	case "EN":
		isCN = false
	case "CN":
		isCN = true
	default:
		isCN = true
	}
	if typeSelect == 1 {
		gs.Download(isCN, downTypeUI[typeSelect]+"_"+input)
	} else if typeSelect == 2 {
		sr.Download(isCN, downTypeUI[typeSelect]+"_"+input)
	}
	fmt.Println(downTypeUI[typeSelect] + " 下载完成！\n按回车键退出…")
	fmt.Scanln()
}
