package utils

type (
	StarRail struct {
		Game
	}
)

func NewStarRail() *StarRail {
	return &StarRail{}
}

func (s *StarRail) Compare(skip bool) {
	s.NCompare("https://api-launcher-static.mihoyo.com/hkrpg_cn/mdk/launcher/api/resource?channel_id=1&key=6KcVuOkbcqjJomjZ&launcher_id=33&sub_channel_id=1", "https://hkrpg-launcher-static.hoyoverse.com/hkrpg_global/mdk/launcher/api/resource?channel_id=1&key=vplOVX8Vn7cwG8yb&launcher_id=35&sub_channel_id=1", skip)
}
