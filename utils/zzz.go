package utils

type (
	ZZZ struct {
		Game
	}
)

func NewZZZ() *ZZZ {
	return &ZZZ{}
}

func (z *ZZZ) Compare(skip bool) {
	z.NCompare("https://hyp-api.mihoyo.com/hyp/hyp-connect/api/getGamePackages?launcher_id=jGHBHlcOq1&game_ids[]=x6znKlJ0xK", "https://sg-hyp-api.hoyoverse.com/hyp/hyp-connect/api/getGamePackages?launcher_id=VYTpXlbWo8&game_ids[]=U5hbdsT9W7", skip)
}
