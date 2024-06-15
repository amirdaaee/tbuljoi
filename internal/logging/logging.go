package logging

import (
	"strings"

	"github.com/amirdaaee/tbuljoi/internal/settings"
	"github.com/sirupsen/logrus"
)

func SetupLogger() {
	ll, err := logrus.ParseLevel(strings.ToLower(settings.Config().LogLevel))
	if err != nil {
		logrus.Panic(err)
	}
	logrus.SetLevel(ll)
}
