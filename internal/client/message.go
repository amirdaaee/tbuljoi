package client

import (
	"context"

	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/tg"
	"github.com/sirupsen/logrus"
)

func GetMessageById(ctx *Context, id int) *tg.Message {
	req := tg.InputMessageID{
		ID: id,
	}
	messages_cls, err := ctx.Raw.MessagesGetMessages(context.Background(), []tg.InputMessageClass{&req})
	if err != nil {
		panic(err)
	}
	messages := messages_cls.(*tg.MessagesMessages)
	message_cls := messages.Messages[0]
	message := message_cls.(*tg.Message)
	return message
}

func GetRepliedMsg(ctx *Context, msg *types.Message) *tg.Message {
	if val_cls, ok := msg.GetReplyTo(); ok {
		val := val_cls.(*tg.MessageReplyHeader)
		return GetMessageById(ctx, val.ReplyToMsgID)
	}
	return nil
}
func GetInviteLinks(msg *tg.Message) []InviteLink {
	iv_markup := GetInviteLinksFromMarkup(msg)
	content_markup := GetInviteLinksFromContent(msg)
	return append(iv_markup, content_markup...)
}
func GetInviteLinksFromMarkup(msg *tg.Message) []InviteLink {
	rep_cls := msg.ReplyMarkup
	rep, ok := rep_cls.(*tg.ReplyInlineMarkup)
	if !ok {
		return []InviteLink{}
	}
	var urls []InviteLink
	for _, r := range rep.Rows {
		for _, b := range r.Buttons {
			btn, ok := b.(*tg.KeyboardButtonURL)
			if !ok {
				continue
			}
			u := btn.URL
			iv := InviteLink(u)
			if iv.IsJoinDeeplink() {
				urls = append(urls, InviteLink(u))
			}
		}
	}
	return urls
}
func GetInviteLinksFromContent(msg *tg.Message) []InviteLink {
	var urls []InviteLink
	re := getInviteLinkRegexp()
	for _, u := range re.FindAllString(msg.Message, -1) {
		iv := InviteLink(u)
		if iv.IsJoinDeeplink() {
			urls = append(urls, iv)
		}
	}
	return urls
}

func ModifyMessage(ctx *Context, chatID int64, msg *types.Message, newText string, append bool) {
	if append {
		newText = GetMessageById(ctx, msg.ID).Message + "\n" + newText
	}
	chReq := tg.MessagesEditMessageRequest{
		Flags:                msg.Flags,
		InvertMedia:          msg.InvertMedia,
		ID:                   msg.ID,
		Message:              newText,
		ReplyMarkup:          msg.ReplyMarkup,
		Entities:             msg.Entities,
		QuickReplyShortcutID: msg.QuickReplyShortcutID,
	}
	if _, err := ctx.EditMessage(chatID, &chReq); err != nil {
		logrus.WithError(err).Error("error modify message")
	}
}
func ForwardMessage(ctx *Context, FromChatID int64, ToChatID int64, msg *tg.Message) error {
	fwReq := tg.MessagesForwardMessagesRequest{
		Background: true,
		ID:         []int{msg.ID},
	}
	_, err := ctx.ForwardMessages(FromChatID, ToChatID, &fwReq)
	return err
}
