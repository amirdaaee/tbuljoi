package bot

import (
	"github.com/amirdaaee/tbuljoi/internal/client"
)

type ClientLogger struct {
	ctx       *handleCtx
	pref      string
	modifyMsg bool
}

func (l *ClientLogger) log(logFN func(args ...interface{}), msg string, append bool) {
	if l.modifyMsg {
		client.ModifyMessage(l.ctx.clCtx, l.ctx.effChatID, l.ctx.effMsg, l.pref+": "+msg, append)
	}
	logFN(msg)
}
