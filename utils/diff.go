package utils

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type (
	diff struct {
		diffData  map[string][]string
		sourceVer string
		targetVer string
		sourceDir string
		targetDir string
	}
)

func NewDiff() *diff {
	return &diff{}
}

func (d *diff) Init(sourceDir string, targetDir string) {
	d.sourceDir = sourceDir
	d.targetDir = targetDir
	d.sourceVer = d.GetVersion(d.sourceDir)
	d.targetVer = d.GetVersion(d.targetDir)
	d.diffData = d.CompareFolders()
}

func (d *diff) isSubdirectory(parentPath string, childPath string) bool {
	relPath, err := filepath.Rel(parentPath, childPath)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(relPath, "..")
}

func (d *diff) bool2int(ok bool) int {
	if ok {
		return 1
	} else {
		return 0
	}
}

func (d *diff) str2bool(str string) bool {
	if str == "0" {
		return false
	} else {
		return true
	}
}

func (d *diff) trimVer(str string) string {
	return strings.Replace(strings.Replace(str, "GenshinImpact_", "", -1), "YuanShen_", "", -1)
}

func (d *diff) CompareFolders() map[string][]string {
	// 1) 枚举源与目标，顺便去掉版本前缀
	var sourceFiles, sourceFolders []string
	filepath.WalkDir(d.sourceDir, func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(d.sourceDir, path)
		if err != nil {
			return err
		}
		if dir.IsDir() {
			if rel != "." {
				sourceFolders = append(sourceFolders, d.trimVer(rel))
			}
		} else {
			sourceFiles = append(sourceFiles, d.trimVer(rel))
		}
		return nil
	})
	var targetFiles, targetFolders []string
	filepath.WalkDir(d.targetDir, func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(d.targetDir, path)
		if err != nil {
			return err
		}
		if dir.IsDir() {
			if rel != "." {
				targetFolders = append(targetFolders, d.trimVer(rel))
			}
		} else {
			targetFiles = append(targetFiles, d.trimVer(rel))
		}
		return nil
	})

	// 2) 目标集合（O(1) 查询）
	tFolderSet := make(map[string]struct{}, len(targetFolders))
	for _, t := range targetFolders {
		tFolderSet[t] = struct{}{}
	}
	tFileSet := make(map[string]struct{}, len(targetFiles))
	for _, t := range targetFiles {
		tFileSet[t] = struct{}{}
	}

	// 3) 只保留顶层目录的辅助函数
	keepTop := func(list []string, p string) []string {
		// 若 p 已被现有父目录覆盖，则跳过
		for _, ex := range list {
			if d.isSubdirectory(ex, p) {
				return list
			}
		}
		// 移除已在列表里的子目录（被新父目录 p 覆盖）
		out := list[:0]
		for _, ex := range list {
			if !d.isSubdirectory(p, ex) {
				out = append(out, ex)
			}
		}
		return append(out, p)
	}

	// 4) 缺失目录分类
	var finalFolders, rootFolders []string
	for _, sdir := range sourceFolders {
		if _, ok := tFolderSet[sdir]; ok {
			continue
		}
		if strings.HasPrefix(sdir, "Data\\") {
			finalFolders = keepTop(finalFolders, sdir)
		} else {
			rootFolders = keepTop(rootFolders, sdir)
		}
	}

	// 5) 文件：排除被目录覆盖；Data\ 下缺失才加入；根级文件也做存在性检查
	var finalFiles, rootFiles []string
	for _, sfile := range sourceFiles {
		covered := false
		for _, p := range finalFolders {
			if d.isSubdirectory(p, sfile) {
				covered = true
				break
			}
		}
		if !covered {
			for _, p := range rootFolders {
				if d.isSubdirectory(p, sfile) {
					covered = true
					break
				}
			}
		}
		if covered {
			continue
		}

		if strings.HasPrefix(sfile, "Data\\") {
			if strings.HasSuffix(sfile, "PCGameSDK.dll") {
				continue
			}
			if _, ok := tFileSet[sfile]; !ok {
				finalFiles = append(finalFiles, sfile)
			}
		} else {
			if strings.HasSuffix(sfile, ".exe") || strings.HasSuffix(sfile, ".dmp") {
				if strings.HasSuffix(sfile, "UnityCrashHandler64.exe") {
					if _, ok := tFileSet[sfile]; !ok {
						rootFiles = append(rootFiles, sfile)
					}
				}
			} else {
				if _, ok := tFileSet[sfile]; !ok {
					rootFiles = append(rootFiles, sfile)
				}
			}
		}
	}

	return map[string][]string{
		"finalFolders": finalFolders,
		"finalFiles":   finalFiles,
		"rootFolders":  rootFolders,
		"rootFiles":    rootFiles,
	}
}

func (d *diff) GetVersion(path string) string {
	retStr := ""
	_, err := os.Stat(filepath.Join(path, "GenshinImpact_Data"))
	if !os.IsNotExist(err) {
		retStr += "GenshinImpact_"
	}
	_, err = os.Stat(filepath.Join(path, "YuanShen_Data"))
	if !os.IsNotExist(err) {
		retStr += "YuanShen_"
	}
	return retStr
}

func (d *diff) CreateSymlinks() {
	err := make([]error, 0)
	for _, folder := range d.diffData["finalFolders"] {
		err = append(err, os.Symlink(filepath.Join(d.sourceDir, d.sourceVer+folder), filepath.Join(d.targetDir, d.targetVer+folder)))
	}
	for _, file := range d.diffData["finalFiles"] {
		err = append(err, os.Symlink(filepath.Join(d.sourceDir, d.sourceVer+file), filepath.Join(d.targetDir, d.targetVer+file)))
	}
	for _, folder := range d.diffData["rootFolders"] {
		err = append(err, os.Symlink(filepath.Join(d.sourceDir, folder), filepath.Join(d.targetDir, folder)))
	}
	for _, file := range d.diffData["rootFiles"] {
		err = append(err, os.Symlink(filepath.Join(d.sourceDir, file), filepath.Join(d.targetDir, file)))
	}
	for _, e := range err {
		if e != nil {
			fmt.Println(e)
		}
	}
}
