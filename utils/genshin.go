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
	g.NCompare("https://sdk-static.mihoyo.com/hk4e_cn/mdk/launcher/api/resource?launcher_id=17&key=KAtdSsoQ&channel_id=14", "https://sdk-os-static.hoyoverse.com/hk4e_global/mdk/launcher/api/resource?key=gcStgarh&launcher_id=10&sub_channel_id=3", skip)
}
