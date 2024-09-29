package bot

import (
	"fmt"
	"strings"

	"github.com/amirdaaee/tbuljoi/internal/client"
	"github.com/amirdaaee/tbuljoi/internal/db"
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
	effMsg     *types.Message
	orgMsg     *tg.Message
	effChat    types.EffectiveChat
	effChatID  int64
	clCtx      *client.Context
	ctx        *ext.Context
	db         *db.Mongo
	channelsDB *db.ChatsCollection
}

func (hCtx *handleCtx) fill(ctx_ *ext.Context, update *ext.Update) {
	hCtx.effMsg = update.EffectiveMessage
	hCtx.effChat = update.EffectiveChat()
	hCtx.effChatID = hCtx.effChat.GetID()
	hCtx.clCtx = client.NewContext(ctx_, hCtx.effMsg, &hCtx.effChat)
	hCtx.ctx = ctx_
	hCtx.orgMsg = client.GetRepliedMsg(hCtx.clCtx, hCtx.effMsg)
	hCtx.db = &db.Mongo{
		DBUri:  settings.Config().MongoURI,
		DBName: settings.Config().MongoDB,
	}
	hCtx.channelsDB = &db.ChatsCollection{
		Mongo:          hCtx.db,
		CollectionName: "channels",
	}
}
func tgHandle(handler handlerType, forceReplied bool) handlers.CallbackResponse {
	return func(ctx_ *ext.Context, update *ext.Update) error {
		hCtx := handleCtx{}
		hCtx.fill(ctx_, update)
		l_ := logrus.WithField("handler", handler.name).WithField("chat_id", hCtx.effChatID)
		clientLogger := ClientLogger{
			ctx:  &hCtx,
			pref: handler.name,
		}
		// ....
		if forceReplied && hCtx.orgMsg == nil {
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
			var err error
			var chanID int64
			if u.IsJoinDeeplink() {
				chanID, err = client.JoinChannelByDeepLink(hCtx.clCtx, u)
			} else if u.IsResolvelink() {
				chanID, err = client.JoinChannelByResolveLink(hCtx.clCtx, u)
			}
			if err != nil {
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
		mongoCl, err := hCtx.db.GetClient()
		if err != nil {
			cl_.log(l_.Error, fmt.Sprintf("error db connection: %s", err), true)
		} else {
			defer mongoCl.Disconnect(hCtx.ctx)
			if err := hCtx.channelsDB.DepChatAppend(hCtx.ctx, mongoCl, hCtx.effChatID, chanIDList); err != nil {
				cl_.log(l_.Error, fmt.Sprintf("error appending channels to db: %s", err), true)
			}
		}
		return nil
	},
}
var unjoinHandler = handlerType{
	name: "unjoin",
	runFN: func(hCtx *handleCtx, l_ *logrus.Entry, cl_ *ClientLogger) error {
		mongoCl, err := hCtx.db.GetClient()
		if err != nil {
			cl_.log(l_.Error, fmt.Sprintf("error db connection: %s", err), false)
			return nil
		}
		defer mongoCl.Disconnect(hCtx.ctx)
		chatDoc := new(db.ChatsDoc)
		if err := hCtx.channelsDB.GetByChatID(hCtx.ctx, mongoCl, chatDoc, hCtx.effChatID); err != nil {
			cl_.log(l_.Error, fmt.Sprintf("error getting channel from db: %s", err), false)
			return nil
		}
		total := len(chatDoc.DepChatID)
		cl_.log(l_.Info, fmt.Sprintf("total %d", total), false)
		for counter, u := range chatDoc.DepChatID {
			if err := client.LeaveChannelById(hCtx.clCtx, u); err != nil {
				cl_.log(l_.Error, fmt.Sprintf("%d/%d Leave Error (%s)", counter+1, total, err.Error()), true)
			} else {
				cl_.log(l_.Info, fmt.Sprintf("%d/%d Leave Done", counter+1, total), true)
			}
		}
		if err := hCtx.channelsDB.DepChatFlush(hCtx.ctx, mongoCl, hCtx.effChatID); err != nil {
			cl_.log(l_.Error, fmt.Sprintf("Flush dep channels from db: %s", err), true)
		}
		return nil
	},
}
var setAutoForword = handlerType{
	name: "autoforward",
	runFN: func(hCtx *handleCtx, l_ *logrus.Entry, cl_ *ClientLogger) error {
		stateStr := strings.TrimPrefix(hCtx.effMsg.Message.Message, "/af ")
		var state bool
		switch stateStr {
		case "t":
			state = true
		case "f":
			state = false
		default:
			cl_.log(l_.Error, fmt.Sprintf("requred state (%s) not understood", stateStr), true)
			return nil
		}
		mongoCl, err := hCtx.db.GetClient()
		if err != nil {
			cl_.log(l_.Error, fmt.Sprintf("error db connection: %s", err), true)
			return nil
		}
		defer mongoCl.Disconnect(hCtx.ctx)

		if err := hCtx.channelsDB.AutoForwardChange(hCtx.ctx, mongoCl, hCtx.effChatID, state); err != nil {
			cl_.log(l_.Error, fmt.Sprintf("error modifing chat doc in db: %s", err), true)
			return nil
		}
		cl_.log(l_.Info, fmt.Sprintf("af set to: %t", state), true)
		return nil
	},
}
