package bot

import (
	"github.com/amirdaaee/tbuljoi/internal/client"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
	"github.com/sirupsen/logrus"
)

func StartBot() {
	cl, err := client.GetClient()
	if err != nil {
		panic(err)
	}
	disp := cl.Dispatcher
	disp.AddHandlerToGroup(handlers.NewMessage(filters.MessageFilter(filterReqJoin), tgHandle(joinHandler, true)), 1)
	disp.AddHandlerToGroup(handlers.NewMessage(filters.MessageFilter(filterReqUnjoin), tgHandle(unjoinHandler, false)), 1)
	disp.AddHandlerToGroup(handlers.NewMessage(filters.MessageFilter(filterReqFwd), tgHandle(forwToSelfHandler, true)), 1)
	disp.AddHandlerToGroup(handlers.NewMessage(filters.MessageFilter(filterReqFwdArch), tgHandle(forwToArchHandler, true)), 1)
	disp.AddHandlerToGroup(handlers.NewMessage(filters.MessageFilter(filterReqAFSet), tgHandle(setAutoForword, false)), 1)
	logrus.Warn("Starting bot...")
	cl.Idle()
}
