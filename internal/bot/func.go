package bot

import (
	"fmt"

	"github.com/amirdaaee/tbuljoi/internal/client"
	"github.com/celestix/gotgproto/ext"
	"github.com/sirupsen/logrus"
)

func forwToSelf(ctx_ *ext.Context, update *ext.Update) error {
	effMsg := update.EffectiveMessage
	effChat := update.EffectiveChat()
	effChatID := effChat.GetID()
	ctx := client.NewContext(ctx_, effMsg, &effChat)
	// ....
	replied := client.GetRepliedMsg(ctx, effMsg)
	if replied == nil {
		return nil
	}
	// ...
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
func joinManyChannel(ctx_ *ext.Context, update *ext.Update) error {
	effMsg := update.EffectiveMessage
	effChat := update.EffectiveChat()
	effChatID := effChat.GetID()
	ctx := client.NewContext(ctx_, effMsg, &effChat)
	// ....
	replied := client.GetRepliedMsg(ctx, effMsg)
	if replied == nil {
		return nil
	}
	// ...
	urls := client.GetInviteLinks(replied)
	logrus.WithField("urls", urls).Info("start")
	// ...
	var updateMess string
	var chanIDList []int64
	client.ModifyMessage(ctx, effChatID, effMsg, fmt.Sprintf(START_JOINING, len(urls)), false)
	for counter, u := range urls {
		ll := logrus.WithField("url", u)
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
		logrus.WithError(err).Error("failed to archive")
		client.ModifyMessage(ctx, effChatID, effMsg, fmt.Sprintf(FAILED_ARCHIVE, err.Error()), true)
	} else {
		client.ModifyMessage(ctx, effChatID, effMsg, DONE_ARCHIVE, true)
	}
	client.ModifyMessage(ctx, effChatID, effMsg, DONE_ALL, true)
	logrus.Info("all done!")
	return nil
}
