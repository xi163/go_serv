package main

import (
	"math/rand"
	"time"

	"github.com/cwloo/gonet/core/base/task"
	"github.com/cwloo/gonet/core/cb"
	"github.com/cwloo/gonet/logs"
	"github.com/cwloo/gonet/utils/cmd"
	db "github.com/cwloo/gonet/utils/dbwraper"
	"github.com/cwloo/server/src/config"
	"github.com/cwloo/server/src/game_login/handler"

	login "github.com/cwloo/server/src/game_login/server"
	"github.com/cwloo/server/src/global"
)

func init() {
	cmd.InitArgs(func(arg *cmd.ARG) {
		arg.SetConf("config/conf.ini")
		arg.AppendPattern("server", "server", "srv", "svr", "s")
		arg.AppendPattern("rpc", "rpc", "r")
	})
}

type Circular_buffer interface {
	Empty() bool
	Full() bool
	Reserve() int
	Capacity() int
	Size() int
	PushFront(v any)
	PushBack(v any)
	PopFront() any
	PopBack() any
	Front() any
	Back() any
}

type circular_buffer struct {
	first, last int
	size, cap   int
	slice       []any
}

func NewCircular(cap int, v any) Circular_buffer {
	s := &circular_buffer{
		first: 0,
		last:  0,
		size:  0,
		cap:   cap,
		slice: make([]any, cap)}
	for i := 0; i < s.cap; i++ {
		s.slice[i] = v
	}
	return s
}

func (s *circular_buffer) Empty() bool {
	return s.Size() == 0
}

func (s *circular_buffer) Full() bool {
	return s.Capacity() == s.Size()
}

func (s *circular_buffer) Reserve() int {
	return s.Capacity() - s.Size()
}

func (s *circular_buffer) Capacity() int {
	return s.cap
}

func (s *circular_buffer) Size() int {
	return s.size
}

func (s *circular_buffer) PushFront(v any) {
	if s.Full() {
		s.PopBack()
	}
	s.slice[s.first] = v
	s.size++
}

func (s *circular_buffer) PushBack(v any) {
	if s.Full() {
		s.PopFront()
	}
	s.slice[s.last] = v
	s.size++
}

func (s *circular_buffer) PopFront() any {
	v := s.slice[s.first]
	s.slice[s.first]=
	for i := s.first + 1; i < s.last; i++ {
		s.slice[i-1] = s.slice[i]
	}
	s.last--
	s.size--
	return v
}

func (s *circular_buffer) PopBack() any {
	v := s.slice[s.last-1]
	for i := s.last-2; i < s.last; i++ {
	}
	s.size--
	return v
}

func (s *circular_buffer) Front() any {
	return s.slice[s.first]
}

func (s *circular_buffer) Back() any {
	return s.slice[s.last]
}

// game_login.exe  --dir-level=3 --conf-name=deploy\config\conf.ini --log-dir=E:\winshare\server/deploy/log --s=192.168.0.105:9787 --r=192.168.0.105:5235
func main() {
	cmd.ParseArgs()
	config.InitGameLoginConfig(cmd.Conf())
	logs.SetTimezone(logs.Timezone(config.Config.Log.GameLogin.Timezone))
	logs.SetMode(logs.Mode(config.Config.Log.GameLogin.Mode))
	logs.SetStyle(logs.Style(config.Config.Log.GameLogin.Style))
	logs.SetLevel(logs.Level(config.Config.Log.GameLogin.Level))
	logs.Init(config.Config.Log.GameLogin.Dir, global.Exe, 100000000)
	rand.Seed(time.Now().UnixNano())
	task.After(time.Duration(config.Config.Interval+(rand.Int()%50))*time.Second, cb.NewFunctor00(func() {
		handler.ReadConfig()
	}))
	db.Init(
		config.Config.Redis,
		config.Config.Mongo,
		config.Config.Mysql,
		config.Config.Mysql)

	// _, err := redisop.HSetOnlineUidGameInfo(10001, 220, 2201)
	// if err != nil {
	// 	logs.Errorf(err.Error())
	// 	return
	// }
	login.Run(cmd.Id(), config.ServiceName())
	logs.Close()
}
