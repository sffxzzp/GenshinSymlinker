package main

import (
	"GenshinSymlinker/utils"
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
	gameType := utils.DetectGame(sourceDir)
	for !utils.PathExists(targetDir) {
		fmt.Print("请输入游戏「换服包」的路径：")
		reader := bufio.NewReader(os.Stdin)
		targetDir, _ = reader.ReadString('\n')
		targetDir = strings.Trim(strings.TrimSpace(targetDir), "\"")
	}
	if utils.IsDirEmpty(targetDir) {
		fmt.Println("检测到「换服包」目录内容为空，将下载换服包…")
		SophonCN := utils.NewSophonGame(true)
		SophonEN := utils.NewSophonGame(false)
		var gs *utils.Genshin
		var sr *utils.StarRail
		var zzz *utils.ZZZ
		useSophon := false
		if gameType == "Genshin" {
			fmt.Println("正在加载原神版本信息，请稍候…")
			if SophonCN.GameExists(gameType) && SophonEN.GameExists(gameType) {
				useSophon = true
			} else {
				gs = utils.NewGenshin()
				gs.Compare(true)
				fmt.Println("版本：", gs.Version)
			}
		} else if gameType == "StarRail" {
			fmt.Println("正在加载星铁版本信息，请稍候…")
			if SophonCN.GameExists(gameType) && SophonEN.GameExists(gameType) {
				useSophon = true
			} else {
				sr = utils.NewStarRail()
				sr.Compare(true)
				fmt.Println("版本：", sr.Version)
			}
		} else if gameType == "ZZZ" {
			fmt.Println("正在加载绝区零版本信息，请稍候…")
			if SophonCN.GameExists(gameType) && SophonEN.GameExists(gameType) {
				useSophon = true
			} else {
				zzz = utils.NewZZZ()
				zzz.Compare(true)
				fmt.Println("版本：", zzz.Version)
			}
		}
		if useSophon {
			bizCN := utils.GetGameByGameType(gameType, true)
			bizEN := utils.GetGameByGameType(gameType, false)
			next := false
			version := SophonCN.Games[bizCN].Version
			if SophonCN.PreGames[bizCN].Version != "" && SophonEN.PreGames[bizEN].Version != "" {
				version = SophonCN.PreGames[bizCN].Version
				next = utils.IsPreDownload(version)
			}
			SophonCN.GetManifest(gameType, next)
			SophonEN.GetManifest(gameType, next)
			utils.SophonDiff(SophonCN, SophonEN)
			fmt.Println("版本：", version)
		}
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
		if useSophon {
			if isCN {
				SophonCN.Download(targetDir)
			} else {
				SophonEN.Download(targetDir)
			}
		} else {
			if gameType == "Genshin" {
				gs.Download(isCN, targetDir)
			} else if gameType == "StarRail" {
				sr.Download(isCN, targetDir)
			} else if gameType == "ZZZ" {
				zzz.Download(isCN, targetDir)
			}
		}
		fmt.Println("换服包下载完成！")
	}
	fmt.Print("正在创建符号链接…")
	d := utils.NewDiff()
	d.Init(sourceDir, targetDir)
	d.CreateSymlinks()
	fmt.Println("完成！")
	fmt.Println("按回车键退出…")
	os.Stdin.Read(make([]byte, 1))
}
