package main

import (
	"bufio"
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

func pathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("%s 未找到！\n", path)
		return false
	}
	return true
}

func newDiff() *diff {
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
	var sourceFiles, sourceFolders []string
	filepath.WalkDir(d.sourceDir, func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(d.sourceDir, path)
		if err != nil {
			return err
		}
		if dir.IsDir() {
			sourceFolders = append(sourceFolders, relPath)
		} else {
			sourceFiles = append(sourceFiles, relPath)
		}
		return nil
	})
	var targetFiles, targetFolders []string
	filepath.WalkDir(d.targetDir, func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(d.targetDir, path)
		if err != nil {
			return err
		}
		if dir.IsDir() {
			targetFolders = append(targetFolders, relPath)
		} else {
			targetFiles = append(targetFiles, relPath)
		}
		return nil
	})
	var diffFolders []string
	for _, sdir := range sourceFolders {
		if sdir == "." {
			continue
		}
		sdir = d.trimVer(sdir)
		dup := false
		for _, tdir := range targetFolders {
			if tdir == "." {
				continue
			}
			tdir = d.trimVer(tdir)
			if sdir == tdir {
				dup = true
				break
			}
		}
		diffFolders = append(diffFolders, fmt.Sprintf("%s|%d", sdir, d.bool2int(dup)))
	}
	var finalFolders, rootFolders []string
	for _, data := range diffFolders {
		tData := strings.Split(data, "|")
		path, ok := tData[0], d.str2bool(tData[1])
		if ok {
			continue
		} else {
			included := false
			for _, tpath := range finalFolders {
				if d.isSubdirectory(tpath, path) {
					included = true
				}
			}
			if !included {
				if strings.HasPrefix(path, "Data\\") {
					finalFolders = append(finalFolders, path)
				} else {
					rootFolders = append(rootFolders, path)
				}
			}
		}
	}
	var diffFiles, rootFiles []string
	for _, sfile := range sourceFiles {
		sfile = d.trimVer(sfile)
		included := false
		for _, tpath := range finalFolders {
			if d.isSubdirectory(tpath, sfile) {
				included = true
			}
		}
		if !included {
			if strings.HasPrefix(sfile, "Data\\") {
				diffFiles = append(diffFiles, sfile)
			} else {
				if !strings.HasSuffix(sfile, ".exe") {
					rootFiles = append(rootFiles, sfile)
				}
			}
		}
	}
	var finalFiles []string
	for _, tmpfile := range diffFiles {
		included := false
		for _, tfile := range targetFiles {
			tfile := d.trimVer(tfile)
			if tmpfile == tfile {
				included = true
			}
		}
		if !included {
			finalFiles = append(finalFiles, tmpfile)
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
	for _, folder := range d.diffData["finalFolders"] {
		os.Symlink(filepath.Join(d.sourceDir, d.sourceVer+folder), filepath.Join(d.targetDir, d.targetVer+folder))
	}
	for _, file := range d.diffData["finalFiles"] {
		os.Symlink(filepath.Join(d.sourceDir, d.sourceVer+file), filepath.Join(d.targetDir, d.targetVer+file))
	}
	for _, folder := range d.diffData["rootFolders"] {
		os.Symlink(filepath.Join(d.sourceDir, folder), filepath.Join(d.targetDir, folder))
	}
	for _, file := range d.diffData["rootFiles"] {
		os.Symlink(filepath.Join(d.sourceDir, file), filepath.Join(d.targetDir, file))
	}
}

func main() {
	sourceDir := "Genshin Impact game"
	targetDir := "YuanShen"
	for !pathExists(sourceDir) {
		fmt.Print("请输入「源文件夹」的路径（支持拖放）：")
		reader := bufio.NewReader(os.Stdin)
		sourceDir, _ = reader.ReadString('\n')
		sourceDir = strings.Trim(strings.TrimSpace(sourceDir), "\"")
	}
	for !pathExists(targetDir) {
		fmt.Print("请输入「换服包」的路径（支持拖放）：")
		reader := bufio.NewReader(os.Stdin)
		targetDir, _ = reader.ReadString('\n')
		targetDir = strings.Trim(strings.TrimSpace(targetDir), "\"")
	}
	d := newDiff()
	d.Init(sourceDir, targetDir)
	d.CreateSymlinks()
}
