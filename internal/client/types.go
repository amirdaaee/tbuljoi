package client

import (
	"regexp"

	"github.com/amirdaaee/tbuljoi/internal/deeplink"
)

type InviteLink string

func getInviteLinkRegexp() *regexp.Regexp {
	reg, err := regexp.Compile(`(?:t|telegram)\.(?:me|dog)/(joinchat/|\+)?([\w-]+)`)
	if err != nil {
		panic(err)
	}
	return reg
}

func (i *InviteLink) IsDeepink() bool {
	return deeplink.IsDeeplinkLike(string(*i))
}
func (i *InviteLink) IsJoinDeeplink() bool {
	if i.IsDeepink() {
		_, err := deeplink.Expect(string(*i), deeplink.Join)
		if err == nil {
			return true
		}
	}
	return false
}
func (i *InviteLink) IsResolvelink() bool {
	if i.IsDeepink() {
		_, err := deeplink.Expect(string(*i), deeplink.Resolve)
		if err == nil {
			return true
		}
	}
	return false
}

func (i *InviteLink) GetJoinHash() string {
	l, err := deeplink.Expect(string(*i), deeplink.Join)
	if err != nil {
		return ""
	}
	return l.Args.Get("invite")
}
