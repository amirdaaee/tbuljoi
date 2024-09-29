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

type ChannelsCollection struct {
	*Mongo
	CollectionName string
}
type ChannelsDoc struct {
	ID         string  `bson:"_id,omitempty"`
	ChannID    int64   `bson:"ChannelID"`
	DepChannID []int64 `bson:"DepChannID"`
}

func (m *ChannelsCollection) GetChannelsCollection(cl *mongo.Client) *mongo.Collection {
	return cl.Database(m.DBName).Collection(m.CollectionName)
}
func (m *ChannelsCollection) GetByChannID(ctx context.Context, cl *mongo.Client, result *ChannelsDoc, chanID int64) error {
	filter := m.FilterChannID(chanID)
	coll := m.GetChannelsCollection(cl)
	return coll.FindOne(ctx, filter).Decode(result)
}
func (m *ChannelsCollection) Insert(ctx context.Context, cl *mongo.Client, channID int64, depChannID []int64) error {
	channDoc := ChannelsDoc{
		ChannID:    channID,
		DepChannID: depChannID,
	}
	_, err := m.GetChannelsCollection(cl).InsertOne(ctx, channDoc)
	return err
}
func (m *ChannelsCollection) DepChannAppend(ctx context.Context, cl *mongo.Client, channID int64, depChannID []int64) error {
	ll := m.getLogger(channID)
	doc := new(ChannelsDoc)
	if err := m.GetByChannID(ctx, cl, doc, channID); err != nil {
		if err == mongo.ErrNoDocuments {
			ll.Info("channel not found in db. inserting")
			return m.Insert(ctx, cl, channID, depChannID)
		} else {
			return fmt.Errorf("error getting mongo channel doc: %s", err)
		}
	}
	depChatSet := mapset.NewSet(doc.DepChannID...)
	depChatSet.Append(depChannID...)
	return m.depChannUpdate(ctx, cl, channID, depChannID)
}
func (m *ChannelsCollection) DepChannFlush(ctx context.Context, cl *mongo.Client, channID int64) error {
	ll := m.getLogger(channID)
	doc := new(ChannelsDoc)
	if err := m.GetByChannID(ctx, cl, doc, channID); err != nil {
		if err == mongo.ErrNoDocuments {
			ll.Info("channel not found in db. inserting")
			return m.Insert(ctx, cl, channID, []int64{})
		} else {
			return fmt.Errorf("error getting mongo channel doc: %s", err)
		}
	}
	return m.depChannUpdate(ctx, cl, channID, []int64{})

}

// ...
func (m *ChannelsCollection) FilterChannID(chanID int64) *primitive.D {
	return &bson.D{{Key: "ChannelID", Value: chanID}}
}

func (m *ChannelsCollection) getLogger(chanID int64) *logrus.Entry {
	return logrus.WithField("channel-id", chanID)
}
func (m *ChannelsCollection) depChannUpdate(ctx context.Context, cl *mongo.Client, chanID int64, depChannID []int64) error {
	upd := bson.M{"$set": bson.M{"DepChannID": depChannID}}
	_, err := m.GetChannelsCollection(cl).UpdateOne(ctx, m.FilterChannID(chanID), upd)
	return err
}
