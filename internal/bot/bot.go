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
	disp.AddHandlerToGroup(handlers.NewMessage(filters.MessageFilter(filterReqJoin), joinManyChannel), 1)
	disp.AddHandlerToGroup(handlers.NewMessage(filters.MessageFilter(filterReqFwd), forwToSelf), 1)
	logrus.Warn("Starting bot...")
	cl.Idle()
}
