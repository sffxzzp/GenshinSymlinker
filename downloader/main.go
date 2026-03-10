package main

import (
	"fmt"
	"strings"

	"GenshinSymlinker/utils"
)

func main() {
	SophonCN := utils.NewSophonGame(true)
	SophonEN := utils.NewSophonGame(false)
	downTypeUI := map[int]string{1: "原神", 2: "星铁", 3: "绝区零"}
	gameTypeList := map[int]string{1: "Genshin", 2: "StarRail", 3: "ZZZ"}
	var typeSelect int
	for typeSelect != 1 && typeSelect != 2 && typeSelect != 3 {
		fmt.Print("游戏\n1. " + downTypeUI[1] + "\n2. " + downTypeUI[2] + "\n3. " + downTypeUI[3] + "\n\n请选择（1/2/3）：")
		fmt.Scanln(&typeSelect)
	}
	gameType := gameTypeList[typeSelect]
	fmt.Println("正在加载" + downTypeUI[typeSelect] + "版本信息，请稍候…")
	if !(SophonCN.GameExists(gameType) && SophonEN.GameExists(gameType)) {
		fmt.Println("当前选择的游戏类型不存在可下载的版本")
		return
	}
	SophonCN.GetManifest(gameType, false)
	SophonCN.DiffList = SophonCN.FileList
	SophonEN.GetManifest(gameType, false)
	SophonEN.DiffList = SophonEN.FileList
	var input string
	for input != "CN" && input != "EN" {
		fmt.Print("要下载的游戏类型（CN/EN）：")
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
	if isCN {
		SophonCN.Download(downTypeUI[typeSelect]+"_"+input, nil)
	} else {
		SophonEN.Download(downTypeUI[typeSelect]+"_"+input, nil)
	}
	fmt.Println(downTypeUI[typeSelect] + " 下载完成！\n按回车键退出…")
	fmt.Scanln()
}
