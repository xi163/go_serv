package handler

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xi123/libgo/logs"
	"github.com/xi123/libgo/utils/cmd"
	"github.com/cwloo/server/src/config"
	"github.com/cwloo/server/src/global"
)

func UpdateCfg(req *global.UpdateCfgReq) (*global.UpdateCfgResp, bool) {
	config.UpdateConfig(cmd.Conf(), req)
	return &global.UpdateCfgResp{
		ErrCode: 0,
		ErrMsg:  "ok"}, true
}

func GetCfg(req *global.GetCfgReq) (*global.GetCfgResp, bool) {
	return config.GetConfig(req)
}

func QueryCacheFile(md5 string) (*global.FileInfoResp, bool) {
	info, _ := global.FileInfos.Get(md5)
	if info == nil {
		return &global.FileInfoResp{Md5: md5, ErrCode: 5, ErrMsg: "not found"}, false
	}
	return &global.FileInfoResp{
		Uuid:    info.Uuid(),
		File:    info.SrcName(),
		Md5:     md5,
		Now:     info.Now(false),
		Total:   info.Total(false),
		ErrCode: 0,
		ErrMsg:  "ok"}, true
}

func QueryCacheFileDetail(md5 string) (*global.FileDetailResp, bool) {
	resp := &global.FileDetailResp{
		ErrCode: 0,
		ErrMsg:  "ok"}
	global.FileInfos.Do(md5, func(info global.FileInfo) {
		progress := float64(info.Now(false)) / float64(info.Total(false))
		// progress, _ := strconv.ParseFloat(fmt.Sprintf("%f", float64(info.Now(false))/float64(info.Total(false))), 64)
		percent := strings.Join([]string{strconv.FormatFloat(progress*100, 'f', 2, 64), "%"}, "")
		ok, _ := info.Ok(false)
		switch ok {
		case true:
			info.Time(false).Sub(info.DateTime())
			resp.File = &global.FileDetail{
				Uuid:     info.Uuid(),
				Md5:      info.Md5(),
				FileName: info.SrcName(),
				DstName:  info.DstName(),
				YunName:  info.YunName(),
				Now:      info.Now(false),
				Total:    info.Total(false),
				Url:      info.Url(false),
				Create:   info.DateTime().Format("20060102150405"),
				Time:     info.Time(false).Format("20060102150405"),
				Percent:  percent,
				Elapsed:  fmt.Sprintf("%v", info.Time(false).Sub(info.DateTime())),
			}
		default:
			resp.File = &global.FileDetail{
				Uuid:     info.Uuid(),
				Md5:      info.Md5(),
				FileName: info.SrcName(),
				DstName:  info.DstName(),
				YunName:  info.YunName(),
				Now:      info.Now(false),
				Total:    info.Total(false),
				Create:   info.DateTime().Format("20060102150405"),
				Percent:  percent,
			}
		}
	})
	return resp, true
}

func QueryCacheUuidList() (*global.UuidListResp, bool) {
	resp := &global.UuidListResp{
		Uuids:   []string{},
		ErrCode: 0,
		ErrMsg:  "ok"}
	global.Uploaders.Range(func(uuid string, uploader global.Uploader) {
		resp.Uuids = append(resp.Uuids, uuid)
	})
	return resp, true
}

func QueryCacheList() (*global.ListResp, bool) {
	resp := &global.ListResp{
		Uuids:   []string{},
		Files:   []*global.FileDetail{},
		ErrCode: 0,
		ErrMsg:  "ok"}
	global.Uploaders.Range(func(uuid string, uploader global.Uploader) {
		resp.Uuids = append(resp.Uuids, uuid)
	})
	global.FileInfos.Range(func(md5 string, info global.FileInfo) {
		progress := float64(info.Now(false)) / float64(info.Total(false))
		// progress, _ := strconv.ParseFloat(fmt.Sprintf("%f", float64(info.Now(false))/float64(info.Total(false))), 64)
		percent := strings.Join([]string{strconv.FormatFloat(progress*100, 'f', 2, 64), "%"}, "")
		ok, _ := info.Ok(false)
		switch ok {
		case true:
			info.Time(false).Sub(info.DateTime())
			resp.Files = append(resp.Files, &global.FileDetail{
				Uuid:     info.Uuid(),
				Md5:      info.Md5(),
				FileName: info.SrcName(),
				DstName:  info.DstName(),
				YunName:  info.YunName(),
				Now:      info.Now(false),
				Total:    info.Total(false),
				Url:      info.Url(false),
				Create:   info.DateTime().Format("20060102150405"),
				Time:     info.Time(false).Format("20060102150405"),
				Percent:  percent,
				Elapsed:  fmt.Sprintf("%v", info.Time(false).Sub(info.DateTime())),
			})
		default:
			resp.Files = append(resp.Files, &global.FileDetail{
				Uuid:     info.Uuid(),
				Md5:      info.Md5(),
				FileName: info.SrcName(),
				DstName:  info.DstName(),
				YunName:  info.YunName(),
				Now:      info.Now(false),
				Total:    info.Total(false),
				Create:   info.DateTime().Format("20060102150405"),
				Percent:  percent,
			})
		}
	})
	return resp, true
}

func FinishedNum() (c int) {
	global.FileInfos.Range(func(md5 string, info global.FileInfo) {
		switch info.Done(false) {
		case true:
			c += 1
		}
	})
	return
}

func PendingNum() (c int) {
	staFromFile := false
	switch staFromFile {
	case true:
		global.FileInfos.Range(func(md5 string, info global.FileInfo) {
			switch info.Done(false) {
			case false:
				c += 1
			}
		})
	default:
		global.Uploaders.Range(func(uuid string, uploader global.Uploader) {
			c += uploader.Len()
		})
	}
	return
}

func DelCacheFile(delType int, md5 string) {
	switch delType {
	case 1:
		// 1-取消文件上传(移除未决的文件)
		global.FileInfos.RemoveWithCond(md5, func(info global.FileInfo) bool {
			return !info.Done(false)
		}, func(info global.FileInfo) {
			os.Remove(config.Config.File.Upload.Dir + info.DstName())
			uploader, _ := global.Uploaders.Get(info.Uuid())
			uploader.Remove(md5)
			info.Put()
		})
	case 2:
		// 2-移除已上传的文件
		global.FileInfos.RemoveWithCond(md5, func(info global.FileInfo) bool {
			if ok, _ := info.Ok(false); ok {
				return true
			}
			return false
		}, func(info global.FileInfo) {
			os.Remove(config.Config.File.Upload.Dir + info.DstName())
			info.Put()
		})
	}
}

func RemovePendingFile(uuid, md5 string) (msg string, ok bool) {
	global.FileInfos.RemoveWithCond(md5, func(info global.FileInfo) bool {
		if info.Uuid() != uuid {
			logs.Fatalf("error")
		}
		if info.Done(false) {
			logs.Fatalf("error")
		}
		return true
	}, func(info global.FileInfo) {
		msg = strings.Join([]string{"RemovePendingFile\n", info.Uuid(), "\n", info.SrcName(), "[", md5, "]\n", info.DstName(), "\n", info.YunName()}, "")
		os.Remove(config.Config.File.Upload.Dir + info.DstName())
		info.Put()
	})
	ok = msg != ""
	return
}

func RemoveCheckErrFile(uuid, md5 string) (msg string, ok bool) {
	global.FileInfos.RemoveWithCond(md5, func(info global.FileInfo) bool {
		if info.Uuid() != uuid {
			logs.Fatalf("error")
		}
		if !info.Done(false) {
			logs.Fatalf("error")
		}
		ok, _ := info.Ok(false)
		return !ok
	}, func(info global.FileInfo) {
		msg = strings.Join([]string{"RemoveCheckErrFile\n", info.Uuid(), "\n", info.SrcName(), "[", md5, "]\n", info.DstName(), "\n", info.YunName()}, "")
		os.Remove(config.Config.File.Upload.Dir + info.DstName())
		info.Put()
	})
	ok = msg != ""
	return
}

func CheckExpiredFile() {
	global.FileInfos.RangeRemoveWithCond(func(info global.FileInfo) bool {
		if ok, _ := info.Ok(false); ok {
			return time.Since(info.HitTime(false)) >= time.Duration(config.Config.File.Upload.FileExpiredTimeout)*time.Second
		}
		return false
	}, func(info global.FileInfo) {
		// os.Remove(dir_upload + info.DstName())
		info.Put()
	})
}

func CheckPendingUploader() {
	switch config.Config.File.Upload.UseAsync > 0 {
	case true:
		////// 异步
		global.Uploaders.Range(func(_ string, uploader global.Uploader) {
			if time.Since(uploader.Get()) >= time.Duration(config.Config.File.Upload.PendingTimeout)*time.Second {
				uploader.NotifyClose()
			}
		})
	default:
		////// 同步
		global.Uploaders.RangeRemoveWithCond(func(uploader global.Uploader) bool {
			return time.Since(uploader.Get()) >= time.Duration(config.Config.File.Upload.PendingTimeout)*time.Second
		}, func(uploader global.Uploader) {
			uploader.Clear()
			uploader.Put()
		})
	}
}
