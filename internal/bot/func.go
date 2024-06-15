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
	logrus.SetLevel(logrus.InfoLevel)
	// ....
	replied := client.GetRepliedMsg(ctx, effMsg)
	if replied == nil {
		return nil
	}
	// ...
	err := client.ForwardMessage(ctx, effChatID, effChatID, replied)
	client.ModifyMessage(ctx, effChatID, effMsg, "Done!")
	return err
}
func joinManyChannel(ctx_ *ext.Context, update *ext.Update) error {
	effMsg := update.EffectiveMessage
	effChat := update.EffectiveChat()
	effChatID := effChat.GetID()
	ctx := client.NewContext(ctx_, effMsg, &effChat)
	// ....
	logrus.SetLevel(logrus.InfoLevel)
	// ....
	replied := client.GetRepliedMsg(ctx, effMsg)
	if replied == nil {
		return nil
	}
	// ...
	urls := client.GetInviteLinks(replied)
	logrus.WithField("urls", urls).Info("start")
	// ...
	updateMess := ""
	var chanIDList []int64
	for counter, u := range urls {
		ll := logrus.WithField("url", u)
		ll.Info("starting")
		if chanID, err := client.JoinChannel(ctx, u); err != nil {
			ll.Error(err)
			updateMess = updateMess + fmt.Sprintf("%d/%d join error\n", counter+1, len(urls))
		} else {
			chanIDList = append(chanIDList, chanID)
			updateMess = updateMess + fmt.Sprintf("%d/%d join done\n", counter+1, len(urls))
		}
		client.ModifyMessage(ctx, effChatID, effMsg, updateMess)
	}
	if _, err := ctx.ArchiveChats(chanIDList); err != nil {
		logrus.WithError(err).Error("failed to archive")
	}
	client.ModifyMessage(ctx, effChatID, effMsg, "Done!")
	logrus.Info("all done!")
	return nil
}
