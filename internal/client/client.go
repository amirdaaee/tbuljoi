package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/amirdaaee/tbuljoi/internal/settings"
	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/glebarez/sqlite"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

func GetClient() (*gotgproto.Client, error) {
	cfg := settings.Config()
	auth_convert := Conversator{}
	device := telegram.DeviceConfig{}
	device.SetDefaults()
	device.DeviceModel = cfg.DeviceName
	cl_opts := &gotgproto.ClientOpts{
		DisableCopyright: true,
		Session:          sessionMaker.SqlSession(sqlite.Open(cfg.SessionFile)),
		AuthConversator:  auth_convert,
		Middlewares:      []telegram.Middleware{floodwait.NewSimpleWaiter().WithMaxRetries(4).WithMaxWait(time.Duration(cfg.MaxFloodWait) * time.Second), ratelimit.New(rate.Every(500*time.Millisecond), 20)},
		Device:           &device,
	}
	client, err := gotgproto.NewClient(
		cfg.AppID,
		cfg.AppHash,
		gotgproto.ClientTypePhone(cfg.PhoneNumber),
		cl_opts,
	)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func JoinChannelByDeepLink(ctx *Context, link InviteLink) (int64, error) {
	chanID := int64(-1)
	if !link.IsJoinDeeplink() {
		return chanID, fmt.Errorf("not a valid deep link (%s)", link)
	}
	logger := logrus.WithField("url", string(link))
	chan_info_cls, err := GetChanellInfoFromDeepLink(ctx, link)
	if err != nil {
		return chanID, err
	}
	switch v := chan_info_cls.(type) {
	case *tg.ChatInvitePeek:
		logger.Info("ChatInvitePeek")
		id, hash, err := GetChatID(v.Chat)
		if err != nil {
			return chanID, err
		}
		if chanID, err = JoinChannelById(ctx, id, hash); err != nil {
			return chanID, err
		}
	case *tg.ChatInvite:
		logger.Info("ChatInvite")
		if chanID, err = JoinChannelByLink(ctx, link); err != nil {
			return chanID, err
		}
	case *tg.ChatInviteAlready:
		logger.Info("ChatInviteAlready")
		id, hash, err := GetChatID(v.Chat)
		if err != nil {
			return chanID, err
		}
		if chanID, err = JoinChannelById(ctx, id, hash); err != nil {
			return chanID, err
		}
	default:
		return chanID, errors.New("chan_info_cls not recognized")
	}
	logger.Infof("joined. id=%d", chanID)
	return chanID, err

}
func JoinChannelByResolveLink(ctx *Context, link InviteLink) (int64, error) {
	chanID := int64(-1)
	if !link.IsResolvelink() {
		return chanID, fmt.Errorf("not a valid resolve link (%s)", link)
	}
	logger := logrus.WithField("url", string(link))
	chan_info_cls, err := GetChanellInfoFromResolveLink(ctx, link)
	if err != nil {
		return chanID, err
	}
	id := chan_info_cls.GetID()
	hash := chan_info_cls.GetAccessHash()
	if chanID, err = JoinChannelById(ctx, id, hash); err != nil {
		return chanID, err
	}
	logger.Infof("joined. id=%d", chanID)
	return chanID, err

}
