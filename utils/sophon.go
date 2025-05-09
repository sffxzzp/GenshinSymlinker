package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	sync "sync"

	"google.golang.org/protobuf/proto"
)

type (
	SophonGame struct {
		BizDict     map[string]string
		Games       map[string]BranchGame
		PreGames    map[string]BranchGame
		LauncherId  map[string]string
		isCN        bool
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
	ChunkDownResult struct {
		Offset int64
		Data   []byte
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
		PreGames: make(map[string]BranchGame),
		Games:    make(map[string]BranchGame),
		isCN:     isCN,
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
			s.PreGames[biz] = BranchGame{
				PackageId: game.Pre.PackageId,
				Branch:    game.Pre.Branch,
				Password:  game.Pre.Password,
				Version:   game.Pre.Version,
			}
		}
		s.Games[biz] = BranchGame{
			PackageId: game.Main.PackageId,
			Branch:    game.Main.Branch,
			Password:  game.Main.Password,
			Version:   game.Main.Version,
		}
	}
}

func (s *SophonGame) GameExists(gameType string) bool {
	biz := GetGameByGameType(gameType, s.isCN)
	if _, ok := s.Games[biz]; ok {
		return true
	}
	return false
}

func (s *SophonGame) GetManifest(gameType string, next bool) {
	apiBase := "https://%s/downloader/sophon_chunk/api/getBuild?branch=%s&package_id=%s&password=%s"
	host := "sg-downloader-api.hoyoverse.com"
	if s.isCN {
		host = "downloader-api.mihoyo.com"
	}
	biz := GetGameByGameType(gameType, s.isCN)
	url := fmt.Sprintf(apiBase, host, s.Games[biz].Branch, s.Games[biz].PackageId, s.Games[biz].Password)
	if _, ok := s.PreGames[biz]; ok && next {
		url = fmt.Sprintf(apiBase, host, s.PreGames[biz].Branch, s.PreGames[biz].PackageId, s.PreGames[biz].Password)
	}
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

	// 使用并发 10 线程下载
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)
	results := make(chan ChunkDownResult, len(f.Chunks))

	for _, chunk := range f.Chunks {
		wg.Add(1)
		go func(chk *SophonChunk) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			chunkData := ZstdGet(urlPrefix + "/" + chk.Id)
			results <- ChunkDownResult{Offset: chk.Offset, Data: chunkData}
		}(chunk)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		start := result.Offset
		end := result.Offset + int64(len(result.Data))
		copy(buf[start:end], result.Data)
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
	size := int64(0)
	for _, file := range s.DiffList {
		size += file.Size
	}
	fmt.Printf("总大小：%.2fMB\n", float64(size)/1024/1024)
	fmt.Printf("文件数：%d\n", len(s.DiffList))
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
	SophonA.FileList = nil
	SophonB.FileList = nil
}
