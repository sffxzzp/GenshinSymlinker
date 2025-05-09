package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
)

type (
	Resource struct {
		Retcode int    `json:"retcode"`
		Message string `json:"message"`
		Data    struct {
			GamePackages []struct {
				Type struct {
					ID   string `json:"id"`
					Name string `json:"biz"`
				} `json:"game"`
				Game struct {
					Major ResGame `json:"major"`
				} `json:"main"`
				PreGame struct {
					Major ResGame `json:"major"`
				} `json:"pre_download"`
			} `json:"game_packages"`
		} `json:"data"`
	}
	ResGame struct {
		Version string `json:"version"`
		BaseUrl string `json:"res_list_url"`
	}
	PkgFile struct {
		RemoteName string `json:"remoteName"`
		MD5        string `json:"md5"`
		FileSize   int    `json:"fileSize"`
	}
	Game struct {
		Version   string
		baseUrlCN string
		baseUrlEN string
		dFilesCN  []string
		dFilesEN  []string
	}
)

func PathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("%s 未找到！\n", path)
		return false
	}
	return true
}

func IsDirEmpty(path string) bool {
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

func DetectGame(path string) string {
	retStr := ""
	_, err := os.Stat(filepath.Join(path, "GenshinImpact_Data"))
	if !os.IsNotExist(err) {
		retStr += "Genshin"
	}
	_, err = os.Stat(filepath.Join(path, "YuanShen_Data"))
	if !os.IsNotExist(err) {
		retStr += "Genshin"
	}
	_, err = os.Stat(filepath.Join(path, "StarRail_Data"))
	if !os.IsNotExist(err) {
		retStr += "StarRail"
	}
	_, err = os.Stat(filepath.Join(path, "ZenlessZoneZero_Data"))
	if !os.IsNotExist(err) {
		retStr += "ZZZ"
	}
	return retStr
}

func HttpGet(url string) []byte {
	res, err := http.Get(url)
	if err != nil {
		return []byte{}
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return []byte{}
	}
	return data
}

func ZstdGet(url string) []byte {
	count := 0
	for count < 3 {
		res, err := http.Get(url)
		if err != nil {
			continue
		}
		defer res.Body.Close()
		decoder, err := zstd.NewReader(res.Body)
		if err != nil {
			continue
		}
		defer decoder.Close()
		var out bytes.Buffer
		_, err = io.Copy(&out, decoder)
		if err != nil {
			continue
		}
		count++
		return out.Bytes()
	}
	return []byte{}
}

func DownFile(url string, path string) bool {
	res, err := http.Get(url)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	dir := filepath.Dir(path)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return false
	}
	out, err := os.Create(path)
	if err != nil {
		return false
	}
	defer out.Close()
	_, err = io.Copy(out, res.Body)
	return err == nil
}

func changeVer(str string) string {
	if strings.HasPrefix(str, "GenshinImpact_") {
		return strings.Replace(str, "GenshinImpact_", "YuanShen_", -1)
	}
	if strings.HasPrefix(str, "YuanShen_") {
		return strings.Replace(str, "YuanShen_", "GenshinImpact_", -1)
	}
	return str
}

func GetGameByGameType(gameType string, isCN bool) string {
	if gameType == "Genshin" {
		gameType = "hk4e"
	} else if gameType == "StarRail" {
		gameType = "hkrpg"
	} else if gameType == "ZZZ" {
		gameType = "nap"
	} else {
		return ""
	}
	if isCN {
		gameType += "_cn"
	} else {
		gameType += "_global"
	}
	return gameType
}

func IsPreDownload(version string) bool {
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

func (g *Game) handlePkg(pkg *[][]byte) *map[string]string {
	pkgFiles := make(map[string]string)
	for _, f := range *pkg {
		var fInfo PkgFile
		json.Unmarshal(f, &fInfo)
		pkgFiles[fInfo.RemoteName] = fInfo.MD5
	}
	return &pkgFiles
}

func (g *Game) fileDiff(pCN *map[string]string, pEN *map[string]string) (dCN []string, dEN []string) {
	dCN = g.fileDiffRaw(pEN, pCN)
	dEN = g.fileDiffRaw(pCN, pEN)
	return dCN, dEN
}

func (g *Game) fileDiffRaw(a *map[string]string, b *map[string]string) (o []string) {
	for rname, md5 := range *b {
		aMD5 := (*a)[changeVer(rname)]
		if md5 != aMD5 {
			o = append(o, rname)
		}
	}
	o = append(o, "pkg_version")
	return o
}

func (g *Game) NCompare(urlCN, urlEN string, skip bool) {
	var resCN, resEN Resource
	json.Unmarshal(HttpGet(urlCN), &resCN)
	json.Unmarshal(HttpGet(urlEN), &resEN)
	next := false
	if resCN.Data.GamePackages[0].PreGame.Major.Version != "" && !skip {
		next = IsPreDownload(resCN.Data.GamePackages[0].PreGame.Major.Version)
	}
	if next {
		g.Version = resCN.Data.GamePackages[0].PreGame.Major.Version
		g.baseUrlCN = resCN.Data.GamePackages[0].PreGame.Major.BaseUrl
		g.baseUrlEN = resEN.Data.GamePackages[0].PreGame.Major.BaseUrl
	} else {
		g.Version = resCN.Data.GamePackages[0].Game.Major.Version
		g.baseUrlCN = resCN.Data.GamePackages[0].Game.Major.BaseUrl
		g.baseUrlEN = resEN.Data.GamePackages[0].Game.Major.BaseUrl
	}
	pkgCN := bytes.Split(HttpGet(g.baseUrlCN+"/pkg_version"), []byte{'\r', '\n'})
	pkgEN := bytes.Split(HttpGet(g.baseUrlEN+"/pkg_version"), []byte{'\r', '\n'})
	pkgFilesCN := g.handlePkg(&pkgCN)
	pkgFilesEN := g.handlePkg(&pkgEN)
	g.dFilesCN, g.dFilesEN = g.fileDiff(pkgFilesCN, pkgFilesEN)
}

func (g *Game) Download(isCN bool, path string) {
	var filelist []string
	var baseUrl string
	if isCN {
		filelist = g.dFilesCN
		baseUrl = g.baseUrlCN
	} else {
		filelist = g.dFilesEN
		baseUrl = g.baseUrlEN
	}
	fmt.Println("正在下载：")
	fmt.Println("文件数：", len(filelist))
	for _, f := range filelist {
		fmt.Println(f)
		count := 0
		for count < 3 {
			if DownFile(baseUrl+"/"+f, filepath.Join(path, f)) {
				break
			} else {
				fmt.Println("下载出错！文件：" + f)
				count++
			}
		}
	}
}
