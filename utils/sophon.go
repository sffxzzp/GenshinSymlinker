package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/proto"
)

type (
	SophonGame struct {
		BizDict     map[string]string
		Games       map[string]BranchGame
		LauncherId  map[string]string
		isCN        bool
		Pre         bool
		ChunkPrefix string
		FileList    []*SophonChunkFile
		DiffList    []*SophonChunkFile
	}
	BranchGame struct {
		PackageId string
		Branch    string
		Password  string
		Version   string
	}
	BranchInfo struct {
		RetCode int    `json:"retcode"`
		Message string `json:"message"`
		Data    struct {
			GameBranches []struct {
				Game struct {
					ID  string `json:"id"`
					Biz string `json:"biz"`
				}
				Main BranchInfoData `json:"main"`
				Pre  BranchInfoData `json:"pre_download"`
			} `json:"game_branches"`
		} `json:"data"`
	}
	BranchInfoData struct {
		PackageId string `json:"package_id"`
		Branch    string `json:"branch"`
		Password  string `json:"password"`
		Version   string `json:"tag"`
	}
	ChunkInfo struct {
		RetCode int    `json:"retcode"`
		Message string `json:"message"`
		Data    struct {
			BuildId   string          `json:"build_id"`
			Tag       string          `json:"tag"`
			Manifests []ChunkManifest `json:"manifests"`
		} `json:"data"`
	}
	ChunkManifest struct {
		CategoryId   string `json:"category_id"`
		CategoryName string `json:"category_name"`
		Manifest     struct {
			Id               string `json:"id"`
			Checksum         string `json:"checksum"`
			CompressedSize   string `json:"compressed_size"`
			UncompressedSize string `json:"uncompressed_size"`
		}
		ChunkDownload     ChunkDownloadStruct `json:"chunk_download"`
		ManifestDownload  ChunkDownloadStruct `json:"manifest_download"`
		MatchingField     string              `json:"matching_field"`
		Stats             ChunkStatsStruct    `json:"stats"`
		DeduplicatedStats ChunkStatsStruct    `json:"deduplicated_stats"`
	}
	ChunkDownloadStruct struct {
		Encryption  int    `json:"encryption"`
		Password    string `json:"password"`
		Compression int    `json:"compression"`
		UrlPrefix   string `json:"url_prefix"`
		UrlSuffix   string `json:"url_suffix"`
	}
	ChunkStatsStruct struct {
		CompressedSize   string `json:"compressed_size"`
		UncompressedSize string `json:"uncompressed_size"`
		FileCount        string `json:"file_count"`
		ChunkCount       string `json:"chunk_count"`
	}
)

func NewSophonGame(isCN bool) *SophonGame {
	ret := &SophonGame{
		BizDict: map[string]string{
			"nap_cn":       "ZZZ",
			"nap_global":   "ZZZ",
			"hkrpg_cn":     "StarRail",
			"hkrpg_global": "StarRail",
			"hk4e_cn":      "Genshin",
			"hk4e_global":  "Genshin",
		},
		LauncherId: map[string]string{
			"cn": "jGHBHlcOq1",
			"en": "VYTpXlbWo8",
		},
		Games: make(map[string]BranchGame),
		isCN:  isCN,
	}
	ret.GetVersion()
	return ret
}

func (s *SophonGame) GetVersion() {
	apiBase := "https://%s/hyp/hyp-connect/api/%s?launcher_id=%s"
	host := "sg-hyp-api.hoyoverse.com"
	launcherId := s.LauncherId["en"]
	if s.isCN {
		host = "hyp-api.mihoyo.com"
		launcherId = s.LauncherId["cn"]
	}
	url := fmt.Sprintf(apiBase, host, "getGameBranches", launcherId)
	res := HttpGet(url)
	var branchInfo BranchInfo
	json.Unmarshal(res, &branchInfo)
	for _, game := range branchInfo.Data.GameBranches {
		biz := game.Game.Biz
		if _, ok := s.BizDict[biz]; !ok {
			continue
		}
		if game.Pre.PackageId != "" {
			s.Pre = true
			s.Games[biz] = BranchGame{
				PackageId: game.Pre.PackageId,
				Branch:    game.Pre.Branch,
				Password:  game.Pre.Password,
				Version:   game.Pre.Version,
			}
		} else {
			s.Games[biz] = BranchGame{
				PackageId: game.Main.PackageId,
				Branch:    game.Main.Branch,
				Password:  game.Main.Password,
				Version:   game.Main.Version,
			}
		}
	}
}

func (s *SophonGame) getGameByGameType(gameType string) string {
	if gameType == "Genshin" {
		gameType = "hk4e"
	} else if gameType == "StarRail" {
		gameType = "hkrpg"
	} else if gameType == "ZZZ" {
		gameType = "nap"
	} else {
		return ""
	}
	if s.isCN {
		gameType += "_cn"
	} else {
		gameType += "_global"
	}
	return gameType
}

func (s *SophonGame) GetManifest(gameType string) {
	apiBase := "https://%s/downloader/sophon_chunk/api/getBuild?branch=%s&package_id=%s&password=%s"
	host := "sg-downloader-api.hoyoverse.com"
	if s.isCN {
		host = "downloader-api.mihoyo.com"
	}
	biz := s.getGameByGameType(gameType)
	url := fmt.Sprintf(apiBase, host, s.Games[biz].Branch, s.Games[biz].PackageId, s.Games[biz].Password)
	res := HttpGet(url)
	var chunkRes ChunkInfo
	json.Unmarshal(res, &chunkRes)
	for _, manifest := range chunkRes.Data.Manifests {
		if s.ChunkPrefix == "" {
			s.ChunkPrefix = manifest.ChunkDownload.UrlPrefix
		}
		manifestUrl := manifest.ManifestDownload.UrlPrefix + "/" + manifest.Manifest.Id
		manifestFile := ZstdGet(manifestUrl)
		var manifestDec SophonChunkManifest
		proto.Unmarshal(manifestFile, &manifestDec)
		s.FileList = append(s.FileList, manifestDec.Chuncks...)
	}
}

func (f *SophonChunkFile) Download(urlPrefix string, path string) bool {
	buf := make([]byte, f.Size)
	for _, chunk := range f.Chunks {
		chunkData := ZstdGet(urlPrefix + "/" + chunk.Id)
		start := chunk.Offset
		end := chunk.Offset + int64(len(chunkData))
		copy(buf[start:end], chunkData)
	}
	hash := md5.Sum(buf)
	if hex.EncodeToString(hash[:]) == f.Md5 {
		filePath := filepath.Join(path, f.File)
		err := os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			return false
		}
		file, err := os.Create(filePath)
		if err != nil {
			return false
		}
		defer file.Close()
		_, err = file.Write(buf)
		return err == nil
	} else {
		return false
	}
}

func (s *SophonGame) Download(path string) {
	fmt.Println("正在下载：")
	for _, file := range s.DiffList {
		fmt.Println(file.File)
		count := 0
		for count < 3 {
			if file.Download(s.ChunkPrefix, path) {
				break
			}
			count++
		}
	}
}

func SophonDiff(SophonA, SophonB *SophonGame) {
	aMap := make(map[string]string, len(SophonA.FileList))
	for _, f := range SophonA.FileList {
		aMap[changeVer(f.File)] = f.Md5
	}
	bMap := make(map[string]string, len(SophonB.FileList))
	for _, f := range SophonB.FileList {
		bMap[f.File] = f.Md5
	}
	for _, f := range SophonA.FileList {
		if md5, ok := bMap[changeVer(f.File)]; !ok || md5 != f.Md5 {
			SophonA.DiffList = append(SophonA.DiffList, f)
		}
	}
	for _, f := range SophonB.FileList {
		if md5, ok := aMap[f.File]; !ok || md5 != f.Md5 {
			SophonB.DiffList = append(SophonB.DiffList, f)
		}
	}
}
