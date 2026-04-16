package diskspaceservice

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
	eventemitter "github.com/Parallels/prl-devops-service/serviceprovider/eventEmitter"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
)

var (
	globalDiskSpaceService *DiskSpaceService
)

const spaceMonitoringTicker = 10 * time.Minute

// ParallesHomePathFn is injected at startup to avoid an import cycle with serviceprovider.
type ParallesHomePathFn func(ctx basecontext.ApiContext, username string) (string, error)

type DiskSpaceService struct {
	ctx                 basecontext.ApiContext
	listenerCtx         context.Context
	cancelFunc          context.CancelFunc
	isRunning           bool
	mu                  sync.Mutex
	parallelsHomepathFn ParallesHomePathFn
}

func Get(ctx basecontext.ApiContext) *DiskSpaceService {
	if globalDiskSpaceService != nil {
		return globalDiskSpaceService
	}
	return New(ctx)
}

func New(ctx basecontext.ApiContext) *DiskSpaceService {
	if globalDiskSpaceService != nil {
		return globalDiskSpaceService
	}

	listenerCtx, cancelFunc := context.WithCancel(context.Background())
	globalDiskSpaceService = &DiskSpaceService{
		ctx:         ctx,
		listenerCtx: listenerCtx,
		cancelFunc:  cancelFunc,
	}
	return globalDiskSpaceService
}

func (d *DiskSpaceService) Name() string {
	return "disk_space_service"
}

func (d *DiskSpaceService) Start() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.isRunning {
		return
	}

	d.isRunning = true
	d.ctx.LogInfof("[DiskSpace] Starting disk space worker")
	go d.startDiskspaceWorker()
}

func (d *DiskSpaceService) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isRunning {
		return
	}

	d.cancelFunc()
	d.isRunning = false
	d.ctx.LogInfof("[DiskSpace] Stopped disk space worker")
}

func (d *DiskSpaceService) IsRunning() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.isRunning
}

func (d *DiskSpaceService) GetCacheDiskSpace(ctx basecontext.ApiContext, diag *errors.Diagnostics) int64 {
	cacheFolder, err := config.Get().CatalogCacheFolder()
	if err != nil {
		rsp := models.NewFromError(err)
		diag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "CatalogCacheFolder")
		return 0
	}
	diskSpace, err := helpers.GetFreeDiskSpace(cacheFolder)
	if err != nil {
		rsp := models.NewFromError(err)
		diag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "GetFreeDiskSpace")
		return 0
	}
	return diskSpace
}

// SetParallelsHomePathProvider injects the function used to query Parallels home disk space.
// Called from startup after both services are initialized.
func (d *DiskSpaceService) SetParallelsHomePathProvider(fn ParallesHomePathFn) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.parallelsHomepathFn = fn
}

func (d *DiskSpaceService) getParallelsHomepath(ctx basecontext.ApiContext, username string) (string, error) {
	if d.parallelsHomepathFn == nil {
		d.ctx.LogErrorf("[DiskSpace] Parallels home path provider function is not set, returning empty path")
		return "", nil
	}
	return d.parallelsHomepathFn(ctx, username)
}

func (d *DiskSpaceService) GetDiskSpaceAvailable(ctx basecontext.ApiContext, username, folderPath string, diag *errors.Diagnostics) models.DiskSpaceAvailable {
	response := models.DiskSpaceAvailable{}

	cacheDiskSpace := d.GetCacheDiskSpace(ctx, diag)
	if diag.HasErrors() {
		return response
	}
	response.CacheFolder = cacheDiskSpace

	if folderPath != "" {
		diskSpace, err := helpers.GetFreeDiskSpace(folderPath)
		if err != nil {
			rsp := models.NewFromError(err)
			diag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "GetFreeDiskSpace")
			return response
		}
		response.Given = diskSpace
	}
	if username == "" {
		if user, err := system.Get().GetCurrentUser(ctx); err == nil {
			ctx.LogInfof("No username provided, using current user %s for disk space info", user)
			username = user
		}
	}

	parallelsHomeDir, err := d.getParallelsHomepath(ctx, username)
	if err != nil {
		rsp := models.NewFromError(err)
		diag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "getParallelsHomepath")
		return response
	}
	parallelsHomeDiskSpace, err := helpers.GetFreeDiskSpace(parallelsHomeDir)
	if err != nil {
		rsp := models.NewFromError(err)
		diag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "GetFreeDiskSpace")
		return response
	}
	response.ParallelsHome = parallelsHomeDiskSpace
	response.PrlHomePath = folderPath
	response.PrlHomePath = parallelsHomeDir
	return response
}

func (d *DiskSpaceService) startDiskspaceWorker() {
	ticker := time.NewTicker(spaceMonitoringTicker)
	defer ticker.Stop()

	d.CheckDiskSpaceAndBroadcast()

	for {
		select {
		case <-d.listenerCtx.Done():
			d.ctx.LogInfof("[DiskSpace] [worker] Stopping disk space worker as listener context is done")
			return
		case <-ticker.C:
			d.CheckDiskSpaceAndBroadcast()
		}
	}
}

func (d *DiskSpaceService) CheckDiskSpaceAndBroadcast() {
	ee := eventemitter.Get()
	if ee == nil || !ee.IsRunning() {
		return
	}

	event := models.DiskSpaceAvailable{}
	workerDiag := errors.NewDiagnostics("BackgroundWorker")
	cacheDiskSpace := d.GetCacheDiskSpace(d.ctx, workerDiag)
	if workerDiag.HasErrors() {
		d.ctx.LogErrorf("[DiskSpace] [worker] Error getting cache disk space: %v", workerDiag.Errors[0].Message)
	} else {
		d.ctx.LogInfof("[DiskSpace] [worker] Cache disk space available: %d MB", cacheDiskSpace)
		event.CacheFolder = cacheDiskSpace
	}

	parallelsHomeDiskSpace, err := d.getParallelsHomepath(d.ctx, "")
	if err != nil {
		d.ctx.LogErrorf("[DiskSpace] [worker] Error getting Parallels home disk space: %v", err)
	} else {
		disksize, err := helpers.GetFreeDiskSpace(parallelsHomeDiskSpace)
		if err != nil {
			d.ctx.LogErrorf("[DiskSpace] [worker] Error getting Parallels home disk space: %v", err)
		} else {
			d.ctx.LogInfof("[DiskSpace] [worker] Parallels home disk space available: %d MB", disksize)
			event.ParallelsHome = disksize
		}
	}

	if err := ee.BroadcastMessage(models.NewEventMessage(constants.EventTypeStats, "DISK_SPACE_CHANGED", event)); err != nil {
		d.ctx.LogErrorf("[DiskSpace] [worker] Error broadcasting disk space event: %v", err)
	}
}
