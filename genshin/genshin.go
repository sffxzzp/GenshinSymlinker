package genshin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type (
	Resource struct {
		Retcode int    `json:"retcode"`
		Message string `json:"message"`
		Data    struct {
			Game struct {
				Latest ResGame `json:"latest"`
			} `json:"game"`
			PreGame struct {
				Latest ResGame `json:"latest"`
			} `json:"pre_download_game"`
		} `json:"data"`
	}
	ResGame struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		BaseUrl string `json:"decompressed_path"`
	}
	PkgFile struct {
		RemoteName string `json:"remoteName"`
		MD5        string `json:"md5"`
		FileSize   int    `json:"fileSize"`
	}
	Genshin struct {
		Version   string
		baseUrlCN string
		baseUrlEN string
		dFilesCN  []string
		dFilesEN  []string
	}
)

func New() *Genshin {
	return &Genshin{}
}

func httpGet(url string) []byte {
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

func changeVer(str string) string {
	if strings.HasPrefix(str, "GenshinImpact_") {
		return strings.Replace(str, "GenshinImpact_", "YuanShen_", -1)
	}
	if strings.HasPrefix(str, "YuanShen_") {
		return strings.Replace(str, "YuanShen_", "GenshinImpact_", -1)
	}
	return str
}

func (g *Genshin) handlePkg(pkg *[][]byte) *map[string]string {
	pkgFiles := make(map[string]string)
	for _, f := range *pkg {
		var fInfo PkgFile
		json.Unmarshal(f, &fInfo)
		pkgFiles[fInfo.RemoteName] = fInfo.MD5
	}
	return &pkgFiles
}

func (g *Genshin) fileDiff(pCN *map[string]string, pEN *map[string]string) (dCN []string, dEN []string) {
	dCN = g.fileDiffRaw(pEN, pCN)
	dEN = g.fileDiffRaw(pCN, pEN)
	return dCN, dEN
}

func (g *Genshin) fileDiffRaw(a *map[string]string, b *map[string]string) (o []string) {
	for rname, md5 := range *b {
		aMD5 := (*a)[changeVer(rname)]
		if md5 != aMD5 {
			o = append(o, rname)
		}
	}
	o = append(o, "pkg_version")
	return o
}

func (g *Genshin) downFile(url string, path string) bool {
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

func (g *Genshin) Compare(skip bool) {
	var resCN, resEN Resource
	json.Unmarshal(httpGet("https://sdk-static.mihoyo.com/hk4e_cn/mdk/launcher/api/resource?launcher_id=17&key=KAtdSsoQ&channel_id=14"), &resCN)
	json.Unmarshal(httpGet("https://sdk-os-static.hoyoverse.com/hk4e_global/mdk/launcher/api/resource?key=gcStgarh&launcher_id=10&sub_channel_id=3"), &resEN)
	next := false
	if resCN.Data.PreGame.Latest.Version != "" && !skip {
		var input string
		for strings.ToUpper(input) != "Y" && strings.ToUpper(input) != "N" {
			fmt.Println("检测到版本 " + resCN.Data.PreGame.Latest.Version + " 的预下载包，是否下载下一版本的换服包（y/N）：")
			fmt.Scanln(&input)
		}
		switch strings.ToUpper(input) {
		case "Y":
			next = true
		case "N":
			next = false
		default:
			next = false
		}
	}
	if next {
		g.Version = resCN.Data.PreGame.Latest.Version
		g.baseUrlCN = resCN.Data.PreGame.Latest.BaseUrl
		g.baseUrlEN = resEN.Data.PreGame.Latest.BaseUrl
	} else {
		g.Version = resCN.Data.Game.Latest.Version
		g.baseUrlCN = resCN.Data.Game.Latest.BaseUrl
		g.baseUrlEN = resEN.Data.Game.Latest.BaseUrl
	}
	pkgCN := bytes.Split(httpGet(g.baseUrlCN+"/pkg_version"), []byte{'\r', '\n'})
	pkgEN := bytes.Split(httpGet(g.baseUrlEN+"/pkg_version"), []byte{'\r', '\n'})
	pkgFilesCN := g.handlePkg(&pkgCN)
	pkgFilesEN := g.handlePkg(&pkgEN)
	g.dFilesCN, g.dFilesEN = g.fileDiff(pkgFilesCN, pkgFilesEN)
}

func (g *Genshin) Download(isCN bool, path string) {
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
	for _, f := range filelist {
		fmt.Println(f)
		count := 0
		for count < 3 {
			if g.downFile(baseUrl+"/"+f, filepath.Join(path, f)) {
				break
			} else {
				fmt.Println("下载出错！文件：" + f)
				count++
			}
		}
	}
}
