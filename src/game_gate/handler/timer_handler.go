package handler

import (
	"time"

	"github.com/cwloo/gonet/core/base/task"
	"github.com/cwloo/gonet/core/cb"
	"github.com/cwloo/gonet/utils/cmd"
	"github.com/cwloo/server/src/config"
)

func ReadConfig() {
	config.ReadConfig(cmd.Conf())
	task.After(time.Duration(config.Config.Interval)*time.Second, cb.NewFunctor00(func() {
		ReadConfig()
	}))
}
