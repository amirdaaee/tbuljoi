package bot

import (
	"fmt"

	"github.com/amirdaaee/tbuljoi/internal/client"
	"github.com/amirdaaee/tbuljoi/internal/settings"
	"github.com/celestix/gotgproto/ext"
	"github.com/sirupsen/logrus"
)

func forwToSelf(ctx_ *ext.Context, update *ext.Update) error {
	effMsg := update.EffectiveMessage
	effChat := update.EffectiveChat()
	effChatID := effChat.GetID()
	ctx := client.NewContext(ctx_, effMsg, &effChat)
	// ....
	l1 := logrus.WithField("chat id", effChatID)
	replied := client.GetRepliedMsg(ctx, effMsg)
	if replied == nil {
		l1.Info("no replied message")
		return nil
	}
	// ...
	l1.Info("started forwarding to self")
	client.ModifyMessage(ctx, effChatID, effMsg, fmt.Sprintf(START_FORWARDING, effChatID), false)
	err := client.ForwardMessage(ctx, effChatID, effChatID, replied)
	if err != nil {
		client.ModifyMessage(ctx, effChatID, effMsg, fmt.Sprintf(FAILED_FORWARDING, effChatID, err), true)
	} else {
		client.ModifyMessage(ctx, effChatID, effMsg, fmt.Sprintf(DONE_FORWARDING, effChatID), true)
	}
	client.ModifyMessage(ctx, effChatID, effMsg, DONE_ALL, true)
	return err
}
func forwToArchive(ctx_ *ext.Context, update *ext.Update) error {
	effMsg := update.EffectiveMessage
	effChat := update.EffectiveChat()
	effChatID := effChat.GetID()
	ctx := client.NewContext(ctx_, effMsg, &effChat)
	// ....
	l1 := logrus.WithField("chat id", effChatID)
	replied := client.GetRepliedMsg(ctx, effMsg)
	if replied == nil {
		l1.Info("no replied message")
		return nil
	}
	archChatID := settings.Config().ArchiveChatID
	// ...
	l1.Info("started forwarding to archive")
	client.ModifyMessage(ctx, effChatID, effMsg, fmt.Sprintf(START_FORWARDING, archChatID), false)
	err := client.ForwardMessage(ctx, effChatID, archChatID, replied)
	if err != nil {
		client.ModifyMessage(ctx, effChatID, effMsg, fmt.Sprintf(FAILED_FORWARDING, archChatID, err), true)
	} else {
		client.ModifyMessage(ctx, effChatID, effMsg, fmt.Sprintf(DONE_FORWARDING, archChatID), true)
	}
	client.ModifyMessage(ctx, effChatID, effMsg, DONE_ALL, true)
	return err
}
func joinManyChannel(ctx_ *ext.Context, update *ext.Update) error {
	effMsg := update.EffectiveMessage
	effChat := update.EffectiveChat()
	effChatID := effChat.GetID()
	ctx := client.NewContext(ctx_, effMsg, &effChat)
	// ....
	l1 := logrus.WithField("chat id", effChatID)
	replied := client.GetRepliedMsg(ctx, effMsg)
	if replied == nil {
		l1.Info("no replied message")
		return nil
	}
	// ...
	urls := client.GetInviteLinks(replied)
	// ...
	var updateMess string
	var chanIDList []int64
	l1.WithField("urls", urls).Info("start")
	client.ModifyMessage(ctx, effChatID, effMsg, fmt.Sprintf(START_JOINING, len(urls)), false)
	for counter, u := range urls {
		ll := l1.WithField("url", u)
		ll.Info("starting")
		if chanID, err := client.JoinChannel(ctx, u); err != nil {
			ll.Error(err)
			updateMess = fmt.Sprintf(FAILED_JOIN, counter+1, len(urls), err.Error())
		} else {
			chanIDList = append(chanIDList, chanID)
			updateMess = fmt.Sprintf(DONE_JOINING, counter+1, len(urls))
		}
		client.ModifyMessage(ctx, effChatID, effMsg, updateMess, true)
	}
	if _, err := ctx.ArchiveChats(chanIDList); err != nil {
		l1.WithError(err).Error("failed to archive")
		client.ModifyMessage(ctx, effChatID, effMsg, fmt.Sprintf(FAILED_ARCHIVE, err.Error()), true)
	} else {
		client.ModifyMessage(ctx, effChatID, effMsg, DONE_ARCHIVE, true)
	}
	client.ModifyMessage(ctx, effChatID, effMsg, DONE_ALL, true)
	l1.Info("all done!")
	return nil
}
