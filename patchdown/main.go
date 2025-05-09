package main

import (
	"fmt"
	"strings"

	"GenshinSymlinker/utils"
)

func main() {
	SophonCN := utils.NewSophonGame(true)
	SophonEN := utils.NewSophonGame(false)
	var gs *utils.Genshin
	var sr *utils.StarRail
	var zzz *utils.ZZZ
	useSophon := false
	downTypeUI := map[int]string{1: "原神", 2: "星铁", 3: "绝区零"}
	gameTypeList := map[int]string{1: "Genshin", 2: "StarRail", 3: "ZZZ"}
	var typeSelect int
	for typeSelect != 1 && typeSelect != 2 && typeSelect != 3 {
		fmt.Print("游戏\n1. " + downTypeUI[1] + "\n2. " + downTypeUI[2] + "\n3. " + downTypeUI[3] + "\n\n请选择（1/2/3）：")
		fmt.Scanln(&typeSelect)
	}
	gameType := gameTypeList[typeSelect]
	if typeSelect == 1 {
		fmt.Println("正在加载" + downTypeUI[typeSelect] + "版本信息，请稍候…")
		if SophonCN.GameExists(gameType) && SophonEN.GameExists(gameType) {
			useSophon = true
		} else {
			gs = utils.NewGenshin()
			gs.Compare(false)
		}
	} else if typeSelect == 2 {
		fmt.Println("正在加载" + downTypeUI[typeSelect] + "版本信息，请稍候…")
		if SophonCN.GameExists(gameType) && SophonEN.GameExists(gameType) {
			useSophon = true
		} else {
			sr = utils.NewStarRail()
			sr.Compare(false)
		}
	} else if typeSelect == 3 {
		fmt.Println("正在加载" + downTypeUI[typeSelect] + "版本信息，请稍候…")
		if SophonCN.GameExists(gameType) && SophonEN.GameExists(gameType) {
			useSophon = true
		} else {
			zzz = utils.NewZZZ()
			zzz.Compare(false)
		}
	}
	if useSophon {
		SophonCN.GetManifest(gameType, false)
		SophonEN.GetManifest(gameType, false)
		utils.SophonDiff(SophonCN, SophonEN)
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
	if useSophon {
		if isCN {
			SophonCN.Download(downTypeUI[typeSelect] + "_" + input)
		} else {
			SophonEN.Download(downTypeUI[typeSelect] + "_" + input)
		}
	} else {
		if typeSelect == 1 {
			gs.Download(isCN, downTypeUI[typeSelect]+"_"+input)
		} else if typeSelect == 2 {
			sr.Download(isCN, downTypeUI[typeSelect]+"_"+input)
		} else if typeSelect == 3 {
			zzz.Download(isCN, downTypeUI[typeSelect]+"_"+input)
		}
	}
	fmt.Println(downTypeUI[typeSelect] + " 下载完成！\n按回车键退出…")
	fmt.Scanln()
}
