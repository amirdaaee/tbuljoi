package bot

import (
	"fmt"

	"github.com/amirdaaee/tbuljoi/internal/client"
	"github.com/amirdaaee/tbuljoi/internal/settings"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/tg"
	"github.com/sirupsen/logrus"
)

type handlerType struct {
	name  string
	runFN func(hCtx *handleCtx, l_ *logrus.Entry, cl_ *ClientLogger) error
}
type handleCtx struct {
	effMsg    *types.Message
	orgMsg    *tg.Message
	effChat   types.EffectiveChat
	effChatID int64
	clCtx     *client.Context
	ctx       *ext.Context
}

func (hCtx *handleCtx) fill(ctx_ *ext.Context, update *ext.Update) {
	hCtx.effMsg = update.EffectiveMessage
	hCtx.effChat = update.EffectiveChat()
	hCtx.effChatID = hCtx.effChat.GetID()
	hCtx.clCtx = client.NewContext(ctx_, hCtx.effMsg, &hCtx.effChat)
	hCtx.ctx = ctx_
	hCtx.orgMsg = client.GetRepliedMsg(hCtx.clCtx, hCtx.effMsg)
}
func tgHandle(handler handlerType) handlers.CallbackResponse {
	return func(ctx_ *ext.Context, update *ext.Update) error {
		hCtx := handleCtx{}
		hCtx.fill(ctx_, update)
		l_ := logrus.WithField("handler", handler.name).WithField("chat_id", hCtx.effChatID)
		clientLogger := ClientLogger{
			ctx:  &hCtx,
			pref: handler.name,
		}
		// ....
		if hCtx.orgMsg == nil {
			clientLogger.log(l_.Warn, "no replied message", false)
			return nil
		}
		// ...
		clientLogger.log(l_.Info, "Start", false)
		err := handler.runFN(&hCtx, l_, &clientLogger)
		if err != nil {
			clientLogger.log(l_.Error, fmt.Sprintf("Error (%s)", err.Error()), true)
		} else {
			clientLogger.log(l_.Info, "Done", true)
		}
		return err
	}
}

var forwToSelfHandler = handlerType{
	name: "forw-to-self",
	runFN: func(hCtx *handleCtx, l_ *logrus.Entry, cl_ *ClientLogger) error {
		return client.ForwardMessage(hCtx.clCtx, hCtx.effChatID, hCtx.effChatID, hCtx.orgMsg)
	},
}

var forwToArchHandler = handlerType{
	name: "forw-to-arch",
	runFN: func(hCtx *handleCtx, l_ *logrus.Entry, cl_ *ClientLogger) error {
		return client.ForwardMessage(hCtx.clCtx, hCtx.effChatID, settings.Config().ArchiveChatID, hCtx.orgMsg)
	},
}

var joinHandler = handlerType{
	name: "join",
	runFN: func(hCtx *handleCtx, l_ *logrus.Entry, cl_ *ClientLogger) error {
		urls := client.GetInviteLinks(hCtx.orgMsg)
		var chanIDList []int64
		cl_.log(l_.Info, fmt.Sprintf("total %d", len(urls)), true)
		for counter, u := range urls {
			if chanID, err := client.JoinChannel(hCtx.clCtx, u); err != nil {
				cl_.log(l_.Error, fmt.Sprintf("%d/%d Join Error (%s)", counter+1, len(urls), err.Error()), true)
			} else {
				chanIDList = append(chanIDList, chanID)
				cl_.log(l_.Info, fmt.Sprintf("%d/%d Join Done", counter+1, len(urls)), true)
			}
		}

		if _, err := hCtx.clCtx.ArchiveChats(chanIDList); err != nil {
			cl_.log(l_.Error, fmt.Sprintf("Archive Error (%s)", err.Error()), true)
		} else {
			cl_.log(l_.Info, "Archive Done", true)
		}
		return nil
	},
}
