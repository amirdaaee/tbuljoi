package client

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/types"
)

type Context struct {
	// original context of the client.
	ext.Context

	EffectiveMessage *types.Message
	EffectiveChat    *types.EffectiveChat
}

func NewContext(ctx *ext.Context, EffectiveMessage *types.Message, EffectiveChat *types.EffectiveChat) *Context {
	return &Context{
		Context:          *ctx,
		EffectiveMessage: EffectiveMessage,
		EffectiveChat:    EffectiveChat,
	}
}
