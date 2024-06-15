package client

import (
	"errors"
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
	device.DeviceModel = "tbuljoi"
	cl_opts := &gotgproto.ClientOpts{
		DisableCopyright: true,
		Session:          sessionMaker.SqlSession(sqlite.Open(cfg.SessionFile)),
		AuthConversator:  auth_convert,
		Middlewares:      []telegram.Middleware{floodwait.NewSimpleWaiter().WithMaxRetries(4), ratelimit.New(rate.Every(500*time.Millisecond), 5)},
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

func JoinChannel(ctx *Context, link InviteLink) (int64, error) {
	chanID := int64(-1)
	if !link.IsJoinDeeplink() {
		return chanID, errors.New("not a valid link")
	}
	logger := logrus.WithField("url", string(link))
	chan_info_cls, err := GetChanellInfoFromLink(ctx, link)
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
