package utils

type (
	Genshin struct {
		Game
	}
)

func NewGenshin() *Genshin {
	return &Genshin{}
}

func (g *Genshin) Compare(skip bool) {
	g.NCompare("https://hyp-api.mihoyo.com/hyp/hyp-connect/api/getGamePackages?launcher_id=jGHBHlcOq1&game_ids[]=1Z8W5NHUQb", "https://sg-hyp-api.hoyoverse.com/hyp/hyp-connect/api/getGamePackages?launcher_id=VYTpXlbWo8&game_ids[]=gopR6Cufr3", skip)
}
