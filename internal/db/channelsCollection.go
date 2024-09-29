package db

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChatsCollection struct {
	*Mongo
	CollectionName string
}
type ChatsDoc struct {
	ID        string  `bson:"_id,omitempty"`
	ChatID    int64   `bson:"ChatID"`
	DepChatID []int64 `bson:"DepChatID"`
}

func (m *ChatsCollection) GetChatsCollection(cl *mongo.Client) *mongo.Collection {
	return cl.Database(m.DBName).Collection(m.CollectionName)
}
func (m *ChatsCollection) GetByChatID(ctx context.Context, cl *mongo.Client, result *ChatsDoc, chatID int64) error {
	filter := m.FilterChatID(chatID)
	coll := m.GetChatsCollection(cl)
	return coll.FindOne(ctx, filter).Decode(result)
}
func (m *ChatsCollection) Insert(ctx context.Context, cl *mongo.Client, chatID int64, depChatID []int64) error {
	chatDoc := ChatsDoc{
		ChatID:    chatID,
		DepChatID: depChatID,
	}
	_, err := m.GetChatsCollection(cl).InsertOne(ctx, chatDoc)
	return err
}
func (m *ChatsCollection) DepChatAppend(ctx context.Context, cl *mongo.Client, chatID int64, depChatID []int64) error {
	ll := m.getLogger(chatID)
	doc := new(ChatsDoc)
	if err := m.GetByChatID(ctx, cl, doc, chatID); err != nil {
		if err == mongo.ErrNoDocuments {
			ll.Info("chat not found in db. inserting")
			return m.Insert(ctx, cl, chatID, depChatID)
		} else {
			return fmt.Errorf("error getting mongo chat doc: %s", err)
		}
	}
	depChatSet := mapset.NewSet(doc.DepChatID...)
	depChatSet.Append(depChatID...)
	return m.depChatUpdate(ctx, cl, chatID, depChatID)
}
func (m *ChatsCollection) DepChatFlush(ctx context.Context, cl *mongo.Client, chatID int64) error {
	ll := m.getLogger(chatID)
	doc := new(ChatsDoc)
	if err := m.GetByChatID(ctx, cl, doc, chatID); err != nil {
		if err == mongo.ErrNoDocuments {
			ll.Info("chat not found in db. inserting")
			return m.Insert(ctx, cl, chatID, []int64{})
		} else {
			return fmt.Errorf("error getting mongo chat doc: %s", err)
		}
	}
	return m.depChatUpdate(ctx, cl, chatID, []int64{})

}
func (m *ChatsCollection) AutoForwardSet(ctx context.Context, cl *mongo.Client, chatID int64) error {
	ll := m.getLogger(chatID)
	doc := new(ChatsDoc)
	if err := m.GetByChatID(ctx, cl, doc, chatID); err != nil {
		if err == mongo.ErrNoDocuments {
			ll.Info("chat not found in db. inserting")
			return m.Insert(ctx, cl, chatID, []int64{})
		} else {
			return fmt.Errorf("error getting mongo chat doc: %s", err)
		}
	}
	return m.depChatUpdate(ctx, cl, chatID, []int64{})

}

// ...
func (m *ChatsCollection) FilterChatID(chatID int64) *primitive.D {
	return &bson.D{{Key: "ChatID", Value: chatID}}
}

func (m *ChatsCollection) getLogger(chatID int64) *logrus.Entry {
	return logrus.WithField("chat-id", chatID)
}
func (m *ChatsCollection) depChatUpdate(ctx context.Context, cl *mongo.Client, chatID int64, depChatID []int64) error {
	upd := bson.M{"$set": bson.M{"DepChatID": depChatID}}
	_, err := m.GetChatsCollection(cl).UpdateOne(ctx, m.FilterChatID(chatID), upd)
	return err
}
