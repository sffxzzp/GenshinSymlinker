package workflow

import (
	"GenshinSymlinker/utils"
	"fmt"
)

type Options struct {
	SourceDir   string
	TargetDir   string
	OnEvent      func(Event)
	DecideCN     func() bool
	DecidePre    func(version string) bool
	SkipDownload bool
}

type VersionInfo struct {
	UseSophon bool
	Version   string
}

func emit(opts Options, ev Event) {
	if opts.OnEvent != nil {
		opts.OnEvent(ev)
	}
}

func Run(opts Options) error {
	if opts.SourceDir == "" || !utils.PathExists(opts.SourceDir) {
		emit(opts, Event{Type: EventError, Step: StepValidate, Err: ErrInvalidPath, Message: "source"})
		return ErrInvalidPath
	}
	if opts.TargetDir == "" || !utils.PathExists(opts.TargetDir) {
		emit(opts, Event{Type: EventError, Step: StepValidate, Err: ErrInvalidPath, Message: "target"})
		return ErrInvalidPath
	}

	emit(opts, Event{Type: EventStepStart, Step: StepDetectGame})
	gameType := utils.DetectGame(opts.SourceDir)
	if gameType == "" {
		emit(opts, Event{Type: EventError, Step: StepDetectGame, Err: ErrUnknownGame})
		return ErrUnknownGame
	}
	emit(opts, Event{Type: EventStepEnd, Step: StepDetectGame, Message: gameType})

	if !opts.SkipDownload && utils.IsDirEmpty(opts.TargetDir) {
		emit(opts, Event{Type: EventStepStart, Step: StepVersion})
		versionInfo, sophonCN, sophonEN, gs, sr, zzz := loadVersion(gameType, opts.DecidePre)
		emit(opts, Event{Type: EventStepEnd, Step: StepVersion, Message: versionInfo.Version})

		if opts.DecideCN == nil {
			return fmt.Errorf("missing region decision")
		}
		isCN := opts.DecideCN()
		emit(opts, Event{Type: EventStepStart, Step: StepDownload})
		if versionInfo.UseSophon {
			if isCN {
				sophonCN.Download(opts.TargetDir, func(name string, index int, total int, err error) {
					emit(opts, Event{Type: EventDownload, Step: StepDownload, FileName: name, Index: index, Total: total, Err: err})
				})
			} else {
				sophonEN.Download(opts.TargetDir, func(name string, index int, total int, err error) {
					emit(opts, Event{Type: EventDownload, Step: StepDownload, FileName: name, Index: index, Total: total, Err: err})
				})
			}
		} else {
			if gameType == "Genshin" {
				gs.Download(isCN, opts.TargetDir, func(name string, index int, total int, err error) {
					emit(opts, Event{Type: EventDownload, Step: StepDownload, FileName: name, Index: index, Total: total, Err: err})
				})
			} else if gameType == "StarRail" {
				sr.Download(isCN, opts.TargetDir, func(name string, index int, total int, err error) {
					emit(opts, Event{Type: EventDownload, Step: StepDownload, FileName: name, Index: index, Total: total, Err: err})
				})
			} else if gameType == "ZZZ" {
				zzz.Download(isCN, opts.TargetDir, func(name string, index int, total int, err error) {
					emit(opts, Event{Type: EventDownload, Step: StepDownload, FileName: name, Index: index, Total: total, Err: err})
				})
			}
		}
		emit(opts, Event{Type: EventStepEnd, Step: StepDownload})
	}

	emit(opts, Event{Type: EventStepStart, Step: StepSymlink})
	d := utils.NewDiff()
	d.Init(opts.SourceDir, opts.TargetDir)
	d.CreateSymlinks()
	emit(opts, Event{Type: EventSymlinkDone, Step: StepSymlink})
	emit(opts, Event{Type: EventStepEnd, Step: StepSymlink})
	emit(opts, Event{Type: EventStepEnd, Step: StepDone})
	return nil
}

func loadVersion(gameType string, decidePre func(version string) bool) (VersionInfo, *utils.SophonGame, *utils.SophonGame, *utils.Genshin, *utils.StarRail, *utils.ZZZ) {
	SophonCN := utils.NewSophonGame(true)
	SophonEN := utils.NewSophonGame(false)
	var gs *utils.Genshin
	var sr *utils.StarRail
	var zzz *utils.ZZZ
	useSophon := false
	version := ""

	if gameType == "Genshin" {
		if SophonCN.GameExists(gameType) && SophonEN.GameExists(gameType) {
			useSophon = true
		} else {
			gs = utils.NewGenshin()
			gs.Compare(false, decidePre)
			version = gs.Version
		}
	} else if gameType == "StarRail" {
		if SophonCN.GameExists(gameType) && SophonEN.GameExists(gameType) {
			useSophon = true
		} else {
			sr = utils.NewStarRail()
			sr.Compare(false, decidePre)
			version = sr.Version
		}
	} else if gameType == "ZZZ" {
		if SophonCN.GameExists(gameType) && SophonEN.GameExists(gameType) {
			useSophon = true
		} else {
			zzz = utils.NewZZZ()
			zzz.Compare(false, decidePre)
			version = zzz.Version
		}
	}

	if useSophon {
		bizCN := utils.GetGameByGameType(gameType, true)
		bizEN := utils.GetGameByGameType(gameType, false)
		next := false
		version = SophonCN.Games[bizCN].Version
		if SophonCN.PreGames[bizCN].Version != "" && SophonEN.PreGames[bizEN].Version != "" {
			version = SophonCN.PreGames[bizCN].Version
			next = utils.IsPreDownload(version, decidePre)
		}
		SophonCN.GetManifest(gameType, next)
		SophonEN.GetManifest(gameType, next)
		utils.SophonDiff(SophonCN, SophonEN)
	}

	return VersionInfo{UseSophon: useSophon, Version: version}, SophonCN, SophonEN, gs, sr, zzz
}
