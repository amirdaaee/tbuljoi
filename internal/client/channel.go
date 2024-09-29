package client

import (
	"errors"
	"fmt"
	"strings"

	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func GetChanellInfoFromDeepLink(ctx *Context, link InviteLink) (tg.ChatInviteClass, error) {
	return ctx.Raw.MessagesCheckChatInvite(ctx, link.GetJoinHash())
}
func GetChanellInfoFromResolveLink(ctx *Context, link InviteLink) (types.EffectiveChat, error) {
	split := strings.Split(string(link), "/")
	return ctx.ResolveUsername(split[len(split)-1])
}

func JoinChannelByLink(ctx *Context, link InviteLink) (int64, error) {
	updateCls, err := ctx.Raw.MessagesImportChatInvite(ctx, link.GetJoinHash())
	if err != nil {
		return -1, err
	}
	update := updateCls.(*tg.Updates)
	return getChanIdFromUpdates(update), nil
}
func JoinChannelById(ctx *Context, id int64, hash int64, ignore ...string) (int64, error) {
	updateCls, err := ctx.Raw.ChannelsJoinChannel(ctx, &tg.InputChannel{ChannelID: id, AccessHash: hash})
	if err != nil {
		if !tgerr.Is(err, ignore...) {
			return -1, err
		}
	}
	return getChanIdFromUpdates(updateCls), nil
}
func LeaveChannelById(ctx *Context, id int64, ignore ...string) error {
	chatCls := ctx.PeerStorage.GetInputPeerById(id)
	chat, ok := chatCls.(*tg.InputPeerChannel)
	if !ok {
		return fmt.Errorf("chat is not channel: %T", chatCls)
	}
	_, err := ctx.Raw.ChannelsLeaveChannel(ctx, &tg.InputChannel{ChannelID: chat.ChannelID, AccessHash: chat.AccessHash})
	if err != nil {
		if !tgerr.Is(err, ignore...) {
			return err
		}
	}
	return nil
}

func GetChatID(chat tg.ChatClass) (int64, int64, error) {
	id := int64(-1)
	hash := int64(-1)
	switch v := chat.(type) {
	case *tg.Channel:
		id = v.ID
		hash = v.AccessHash
	default:
		return id, hash, errors.New("unexpexted chat class")
	}
	return id, hash, nil
}

func getChanIdFromUpdates(updateCls tg.UpdatesClass) int64 {
	update := updateCls.(*tg.Updates)
	return update.Chats[0].(*tg.Channel).ID
}
