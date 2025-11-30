package client

import (
	"context"
	"fmt"

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
		logrus.Error(err)
		return nil
	}
	messages := messages_cls.(*tg.MessagesMessages)
	message_cls := messages.Messages[0]
	message, ok := message_cls.(*tg.Message)
	if ok {
		return message
	} else {
		return nil
	}
}

func GetRepliedMsg(ctx *Context, msg *types.Message) *tg.Message {
	if val_cls, ok := msg.GetReplyTo(); ok {
		if val, ok := val_cls.(*tg.MessageReplyHeader); ok {
			return GetMessageById(ctx, val.ReplyToMsgID)
		}
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
			if iv.IsJoinDeeplink() || iv.IsResolvelink() {
				urls = append(urls, InviteLink(u))
			} else {
				logrus.WithField("url", u).Warn("not a deep link")
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
		if iv.IsJoinDeeplink() || iv.IsResolvelink() {
			urls = append(urls, iv)
		} else {
			logrus.WithField("url", u).Warn("not a deep link")
		}
	}
	return urls
}

func ModifyMessage(ctx *Context, chatID int64, msg *types.Message, newText string, append bool) {
	if append {
		oldMsg := GetMessageById(ctx, msg.ID)
		if oldMsg != nil {
			newText = oldMsg.Message + "\n" + newText
		}
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
func AllMediaInChat(ctx *Context, FromChatID int64) ([]*tg.Message, error) {
	// Get InputPeer from chatID
	peer := ctx.PeerStorage.GetInputPeerById(FromChatID)
	if peer == nil {
		logrus.WithField("chat_id", FromChatID).Error("failed to get input peer")
		return nil, fmt.Errorf("failed to get input peer for chat_id: %d", FromChatID)
	}

	messageMap := make(map[int]*tg.Message)
	limit := 100 // Telegram API limit

	photoVideoFilter := &tg.InputMessagesFilterPhotoVideo{}
	err := searchMediaMessages(ctx, peer, photoVideoFilter, messageMap, limit)
	if err != nil {
		logrus.WithError(err).Error("error searching photo/video messages")
	}
	documentFilter := &tg.InputMessagesFilterDocument{}
	err = searchMediaMessages(ctx, peer, documentFilter, messageMap, limit)
	if err != nil {
		logrus.WithError(err).Error("error searching document messages")
	}
	videoNoteFilter := &tg.InputMessagesFilterRoundVideo{}
	err = searchMediaMessages(ctx, peer, videoNoteFilter, messageMap, limit)
	if err != nil {
		logrus.WithError(err).Error("error searching video note messages")
	}

	// Convert map to slice
	allMediaMessages := make([]*tg.Message, 0, len(messageMap))
	for _, msg := range messageMap {
		allMediaMessages = append(allMediaMessages, msg)
	}

	return allMediaMessages, nil
}

func searchMediaMessages(ctx *Context, peer tg.InputPeerClass, filter tg.MessagesFilterClass, results map[int]*tg.Message, limit int) error {
	offsetID := 0
	prevOffsetID := -1

	for {
		req := &tg.MessagesSearchRequest{
			Peer:     peer,
			Q:        "", // Empty query to get all messages
			Filter:   filter,
			OffsetID: offsetID,
			Limit:    limit,
		}

		resp, err := ctx.Raw.MessagesSearch(ctx, req)
		if err != nil {
			return err
		}

		var messages *tg.MessagesMessages
		switch v := resp.(type) {
		case *tg.MessagesMessages:
			messages = v
		case *tg.MessagesMessagesSlice:
			messages = &tg.MessagesMessages{
				Messages: v.Messages,
				Chats:    v.Chats,
				Users:    v.Users,
			}
		default:
			logrus.WithField("type", fmt.Sprintf("%T", v)).Warn("unexpected response type")
			return nil
		}

		if len(messages.Messages) == 0 {
			break
		}

		// Extract messages with media and find minimum ID for pagination
		minID := -1
		for _, msgClass := range messages.Messages {
			msg, ok := msgClass.(*tg.Message)
			if !ok {
				continue
			}
			// Double-check that message has media
			if msg.Media != nil {
				// Deduplicate by message ID
				if _, exists := results[msg.ID]; !exists {
					results[msg.ID] = msg
				}
				// Track minimum ID for pagination (oldest message)
				if minID == -1 || msg.ID < minID {
					minID = msg.ID
				}
			}
		}

		// If no media messages found in this batch, break
		if minID == -1 {
			break
		}

		// Update offset for next iteration (search for older messages)
		prevOffsetID = offsetID
		offsetID = minID

		// If we got fewer messages than requested, we've reached the end
		if len(messages.Messages) < limit {
			break
		}

		// If offsetID didn't change, break to avoid infinite loop
		if offsetID == prevOffsetID {
			break
		}
	}

	return nil
}
