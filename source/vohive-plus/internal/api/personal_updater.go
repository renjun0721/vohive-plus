package api

import (
	"context"
	"github.com/yibaiba/hideck/internal/global"
	"github.com/yibaiba/hideck/internal/updater"
)

type personalUpdateChecker struct{}

func (personalUpdateChecker) CheckUpdate(context.Context) (*updater.UpdateInfo, error) {
	return &updater.UpdateInfo{
		CurrentVer: global.Version, LatestVer: global.Version, HasUpdate: false,
		IsDocker:    updater.DetectDockerEnvironment(),
		ReleaseNote: "个人定制版通过保留的源码构建更新；请勿用官方 HiDeck 镜像覆盖个人版。",
	}, nil
}
