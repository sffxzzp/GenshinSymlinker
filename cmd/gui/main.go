package main

import (
	"GenshinSymlinker/utils"
	"GenshinSymlinker/workflow"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/text"
	"github.com/sqweek/dialog"
)

type Step int

const (
	stepIdle Step = iota
	stepDownload
	stepSymlink
	stepDone
)

type UIState struct {
	step Step

	sourceEditor widget.Editor
	targetEditor widget.Editor

	browseSource widget.Clickable
	browseTarget widget.Clickable

	cnRadio widget.Enum
	preCheck widget.Bool
	gameType string

	startButton widget.Clickable

	status string
	progress string
	progressRatio float32
	lastFile string
	lastError string
	lastVersion string
}

func main() {
	go func() {
		w := new(app.Window)
		w.Option(
			app.Title("GenshinSymlinker"),
			app.Size(unit.Dp(540), unit.Dp(520)),
			app.MinSize(unit.Dp(540), unit.Dp(520)),
			app.MaxSize(unit.Dp(540), unit.Dp(520)),
		)
		if err := run(w); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(w *app.Window) error {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	state := &UIState{step: stepIdle}
	state.cnRadio.Value = "CN"
	state.preCheck.Value = true
	var ops op.Ops

	events := make(chan workflow.Event, 32)
	errCh := make(chan error, 1)

	var sourceDir string
	var targetDir string

	for {
		e := w.Event()
		switch e := e.(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			drainEvents(events, state)
			if state.lastError != "" {
				state.status = state.lastError
			}

			sourceDir = strings.TrimSpace(state.sourceEditor.Text())
			targetDir = strings.TrimSpace(state.targetEditor.Text())
			if sourceDir != "" {
				state.gameType = utils.DetectGame(sourceDir)
			} else {
				state.gameType = ""
			}

			sourceDone := sourceDir != ""
			targetDone := sourceDone && targetDir != ""
			optionsReady := targetDone && state.step == stepIdle

			browseSourceClicked := state.browseSource.Clicked(gtx)
			browseTargetClicked := state.browseTarget.Clicked(gtx)

			if browseSourceClicked {
				if dir, err := dialog.Directory().Title("选择源目录").Browse(); err == nil {
					state.sourceEditor.SetText(dir)
				}
			}
			if browseTargetClicked && sourceDone {
				if dir, err := dialog.Directory().Title("选择目标目录").Browse(); err == nil {
					state.targetEditor.SetText(dir)
				}
			}

			startClicked := state.startButton.Clicked(gtx)

			if startClicked && optionsReady {
				state.status = "正在启动..."
				state.step = stepDownload
				state.status = "准备开始"
				state.lastError = ""
				state.progress = ""
				state.progressRatio = 0
				state.lastFile = ""
				state.lastVersion = ""
				go func() {
					errCh <- workflow.Run(workflow.Options{
						SourceDir: sourceDir,
						TargetDir: targetDir,
						DecideCN: func() bool {
							return state.cnRadio.Value == "CN"
						},
						DecidePre: func(version string) bool {
							return state.preCheck.Value
						},
						OnEvent: func(ev workflow.Event) {
							events <- ev
						},
					})
				}()
			}

			layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return header(th, gtx, state)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min.X = gtx.Constraints.Max.X
								gtx.Constraints.Max.X = gtx.Constraints.Min.X
								return sectionCard(th, gtx, "源目录", func(gtx layout.Context) layout.Dimensions {
									return sourceStep(th, gtx, state)
								})
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								gtxChild := gtx
								if !sourceDone {
									gtxChild = gtxChild.Disabled()
								}
								gtxChild.Constraints.Min.X = gtxChild.Constraints.Max.X
								gtxChild.Constraints.Max.X = gtxChild.Constraints.Min.X
								return sectionCard(th, gtxChild, "目标目录", func(gtx layout.Context) layout.Dimensions {
									return targetStep(th, gtx, state)
								})
							}),
						)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtxChild := gtx
								if !targetDone || state.step != stepIdle {
									gtxChild = gtx.Disabled()
								}
								return sectionCard(th, gtxChild, "选项", func(gtx layout.Context) layout.Dimensions {
									return optionsStep(th, gtx, state)
								})
							}),
							layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								gtxChild := gtx
								if state.step == stepIdle {
									gtxChild = gtx.Disabled()
								}
								return sectionCard(th, gtxChild, "进度", func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceAround}.Layout(gtx,
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											return progressStep(th, gtx, state)
										}),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											lbl := material.Body2(th, state.status)
											lbl.Color = color.NRGBA{R: 0x55, G: 0x55, B: 0x55, A: 0xff}
											return lbl.Layout(gtx)
										}),
									)
								})
							}),
						)
					}),
				)
			})

			select {
			case err := <-errCh:
				if err != nil {
					state.lastError = err.Error() + "，请检查路径或权限（需管理员权限或开发者模式）。"
				}
			default:
			}

			e.Frame(gtx.Ops)
		}
	}
}

func header(th *material.Theme, gtx layout.Context, state *UIState) layout.Dimensions {
	return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		title := "GenshinSymlinker GUI"
		if state.gameType != "" {
			title = title + " - " + state.gameType
		}
		lbl := material.H6(th, title)
		lbl.Color = color.NRGBA{A: 0xff}
		return lbl.Layout(gtx)
	})
}


func sourceStep(th *material.Theme, gtx layout.Context, state *UIState) layout.Dimensions {
	state.sourceEditor.SingleLine = true
	return stepLayout(th, gtx, "源目录", &state.sourceEditor, &state.browseSource)
}

func targetStep(th *material.Theme, gtx layout.Context, state *UIState) layout.Dimensions {
	state.targetEditor.SingleLine = true
	return stepLayout(th, gtx, "目标目录", &state.targetEditor, &state.browseTarget)
}

func stepLayout(th *material.Theme, gtx layout.Context, label string, editor *widget.Editor, browse *widget.Clickable) layout.Dimensions {
	return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceAround}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Editor(th, editor, "选择或输入路径").Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return minButtonHeight(gtx, func(gtx layout.Context) layout.Dimensions {
							return material.Button(th, browse, "浏览").Layout(gtx)
						})
					}),
				)
			}),
		)
	})
}

func optionsStep(th *material.Theme, gtx layout.Context, state *UIState) layout.Dimensions {
	return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceAround}.Layout(gtx,
			layout.Rigid(material.Body2(th, "目标目录为空时会下载换服包").Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Dimensions{}
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							layout.Rigid(material.RadioButton(th, &state.cnRadio, "CN", "CN").Layout),
							layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
							layout.Rigid(material.RadioButton(th, &state.cnRadio, "EN", "EN").Layout),
							layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
							layout.Rigid(material.CheckBox(th, &state.preCheck, "使用预下载版本（如有）").Layout),
						)
					}),
				)
			}),
			layout.Rigid(material.Body2(th, "若创建符号链接失败，请以管理员权限运行或开启开发者模式。").Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return minButtonHeight(gtx, func(gtx layout.Context) layout.Dimensions {
					return material.Button(th, &state.startButton, "开始").Layout(gtx)
				})
			}),
		)
	})
}

func progressStep(th *material.Theme, gtx layout.Context, state *UIState) layout.Dimensions {
	title := "就绪"
	if state.step == stepDownload {
		title = "正在下载"
	} else if state.step == stepSymlink {
		title = "正在创建符号链接"
	} else if state.step == stepDone {
		title = "完成"
	}
	return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		versionText := state.lastVersion
		progressText := state.progress
		fileText := state.lastFile
		if state.step == stepIdle {
			versionText = " "
			progressText = " "
			fileText = " "
		}
		minHeight := gtx.Dp(unit.Dp(120))
		if gtx.Constraints.Max.Y > 0 && gtx.Constraints.Max.Y < minHeight {
			minHeight = gtx.Constraints.Max.Y
		}
		gtx.Constraints.Min.Y = minHeight
		gtx.Constraints.Max.Y = minHeight
		content := layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceAround}
		return content.Layout(gtx,
			layout.Rigid(material.Body2(th, title).Layout),
			layout.Rigid(material.Body2(th, versionText).Layout),
			layout.Rigid(material.Body2(th, progressText).Layout),
			layout.Rigid(material.Body2(th, fileText).Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				bar := material.ProgressBar(th, state.progressRatio)
				bar.Height = unit.Dp(6)
				return bar.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		)
	})
}

func minButtonHeight(gtx layout.Context, widgetLayout layout.Widget) layout.Dimensions {
	gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(40))
	return widgetLayout(gtx)
}

func forceWidth(gtx layout.Context, width int, content layout.Widget) layout.Dimensions {
	gtx.Constraints.Min.X = width
	gtx.Constraints.Max.X = width
	return content(gtx)
}

func spacerPx(px int) layout.Dimensions {
	return layout.Dimensions{Size: image.Pt(px, 0)}
}

func sectionCard(th *material.Theme, gtx layout.Context, title string, content layout.Widget) layout.Dimensions {
	bg := color.NRGBA{R: 0xfa, G: 0xfa, B: 0xfa, A: 0xff}
	border := color.NRGBA{R: 0xe3, G: 0xe3, B: 0xe3, A: 0xff}
	return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		record := op.Record(gtx.Ops)
		dims := layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body2(th, title)
					lbl.Color = color.NRGBA{R: 0x55, G: 0x55, B: 0x55, A: 0xff}
					return lbl.Layout(gtx)
				}),
				layout.Rigid(content),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{}
				}),
			)
		})
		call := record.Stop()
		if dims.Size.X < gtx.Constraints.Max.X {
			dims.Size.X = gtx.Constraints.Max.X
		}
		rect := image.Rectangle{Max: dims.Size}
		rrect := clip.RRect{Rect: rect, NE: 6, NW: 6, SE: 6, SW: 6}
		paint.FillShape(gtx.Ops, bg, rrect.Op(gtx.Ops))
		stroke := clip.Stroke{Path: rrect.Path(gtx.Ops), Width: 1}.Op()
		paint.FillShape(gtx.Ops, border, stroke)
		call.Add(gtx.Ops)
		return dims
	})
}

func drainEvents(events <-chan workflow.Event, state *UIState) {
	for {
		select {
		case ev := <-events:
			switch ev.Type {
			case workflow.EventStepStart:
				if ev.Step == workflow.StepDownload {
					state.status = "开始下载"
					state.step = stepDownload
				}
				if ev.Step == workflow.StepSymlink {
					state.status = "创建符号链接"
					state.step = stepSymlink
				}
			case workflow.EventStepEnd:
				if ev.Step == workflow.StepVersion && ev.Message != "" {
					state.lastVersion = "版本：" + ev.Message
				}
				if ev.Step == workflow.StepDownload {
					state.status = "下载完成"
				}
				if ev.Step == workflow.StepSymlink {
					state.status = "符号链接完成"
					state.step = stepDone
				}
			case workflow.EventDownload:
				if ev.Total > 0 {
					state.progress = "文件 " + itoa(ev.Index) + "/" + itoa(ev.Total)
					state.progressRatio = float32(ev.Index) / float32(ev.Total)
				}
				if ev.FileName != "" {
					state.lastFile = filepath.Base(ev.FileName)
				}
				if ev.Err != nil {
					state.lastError = ev.Err.Error()
				}
			case workflow.EventError:
				if ev.Err != nil {
					state.lastError = ev.Err.Error() + "，请检查路径或权限（需管理员权限或开发者模式）。"
				}
			case workflow.EventSymlinkDone:
				state.status = "完成"
				state.step = stepDone
				state.progressRatio = 1
			}
		default:
			return
		}
	}
}

func itoa(v int) string {
	return strconv.Itoa(v)
}
