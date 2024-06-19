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
	s.NCompare("https://hyp-api.mihoyo.com/hyp/hyp-connect/api/getGamePackages?launcher_id=jGHBHlcOq1&game_ids[]=64kMb5iAWu", "https://sg-hyp-api.hoyoverse.com/hyp/hyp-connect/api/getGamePackages?launcher_id=VYTpXlbWo8&game_ids[]=4ziysqXOQ8", skip)
}
