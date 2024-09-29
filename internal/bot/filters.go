package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/amirdaaee/tbuljoi/internal/db"
	"github.com/amirdaaee/tbuljoi/internal/settings"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/tg"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

func isFromSelf(m *types.Message) bool {
	return m.Out
}
func isAF(m *types.Message) bool {
	afCache := GetAFCache()
	peerIDCls := m.PeerID
	var peerID int64
	switch v := peerIDCls.(type) {
	case *tg.PeerUser:
		peerID = v.UserID
	case *tg.PeerChat:
		peerID = v.ChatID
	case *tg.PeerChannel:
		peerID = v.ChannelID
	default:
		logrus.Errorf("unexpected peer type: %T", v)
		return false
	}
	peerIDStr := fmt.Sprintf("%d", peerID)
	ll := logrus.WithField("peer-id", peerID)
	valCache, found := afCache.Get(peerIDStr)
	var val bool
	if found {
		val = valCache.(bool)
		ll.Debugf("af found in cache: %t", val)
	} else {
		ll.Debug("af not found in cache")
		ctx := context.Background()
		dbStruct := getDB()
		coll := getChatsDBCollection(dbStruct)
		cl, err := dbStruct.GetClient()
		if err != nil {
			logrus.Errorf("error getting client: %s", err)
			return false
		}
		defer cl.Disconnect(ctx)
		res := new(db.ChatsDoc)
		err = coll.GetByChatID(ctx, cl, res, peerID)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				logrus.Errorf("error getting doc from db: %s", err)
				return false
			} else {
				ll.Debugf("no db record found. settign af false")
				val = false
			}
		} else {
			ll.Debugf("db record found. settign af: %+v", res)
			val = res.AutoForward
		}
		afCache.Set(peerIDStr, val, 0)
		ll.Debugf("set af cache: %s:%t", peerIDStr, val)
	}
	return val
}
func isAFRelaxed(m *types.Message) bool {
	afCache := GetAFRelaxCache()
	peerIDCls := m.PeerID
	var peerID int64
	switch v := peerIDCls.(type) {
	case *tg.PeerUser:
		peerID = v.UserID
	case *tg.PeerChat:
		peerID = v.ChatID
	case *tg.PeerChannel:
		peerID = v.ChannelID
	default:
		logrus.Errorf("unexpected peer type: %T", v)
		return false
	}
	peerIDStr := fmt.Sprintf("%d", peerID)
	ll := logrus.WithField("peer-id", peerID)
	valCache, found := afCache.Get(peerIDStr)
	if found {
		val := valCache.(time.Time)
		dur := time.Since(val)
		if dur > settings.Config().AFBurst {
			ll.Debug("relaxing ...")
			return true
		}
		return false
	} else {
		afCache.Set(peerIDStr, time.Now(), 0)
		return false
	}
}

func filterReqJoin(m *types.Message) bool {
	return isFromSelf(m) && m.Text == "/j"
}
func filterReqUnjoin(m *types.Message) bool {
	return isFromSelf(m) && m.Text == "/uj"
}
func filterReqFwd(m *types.Message) bool {
	return isFromSelf(m) && m.Text == "/f"
}
func filterReqFwdArch(m *types.Message) bool {
	return isFromSelf(m) && m.Text == "/ff"
}
func filterReqAFSet(m *types.Message) bool {
	return isFromSelf(m) && strings.HasPrefix(m.Text, "/af ")
}
func filterAF(m *types.Message) bool {
	return !isFromSelf(m) && isAF(m) && !isAFRelaxed(m)
}
