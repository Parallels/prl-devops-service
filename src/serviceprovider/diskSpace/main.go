package diskspaceservice

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
	eventemitter "github.com/Parallels/prl-devops-service/serviceprovider/eventEmitter"
)

var (
	globalDiskSpaceService *DiskSpaceService
)

const spaceMonitoringTicker = 10 * time.Minute

// HomeDiskSpaceFn is injected at startup to avoid an import cycle with serviceprovider.
type HomeDiskSpaceFn func(ctx basecontext.ApiContext, username string) (int64, error)

type DiskSpaceService struct {
	ctx             basecontext.ApiContext
	listenerCtx     context.Context
	cancelFunc      context.CancelFunc
	isRunning       bool
	mu              sync.Mutex
	homeDiskSpaceFn HomeDiskSpaceFn
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

func (d *DiskSpaceService) GetCacheDiskSpace(ctx basecontext.ApiContext) (int64, error) {
	cacheFolder, err := config.Get().CatalogCacheFolder()
	if err != nil {
		return 0, err
	}
	diskSpace, err := helpers.GetFreeDiskSpace(cacheFolder)
	if err != nil {
		return 0, err
	}
	return diskSpace, nil
}

// SetHomeDiskSpaceProvider injects the function used to query Parallels home disk space.
// Called from startup after both services are initialised.
func (d *DiskSpaceService) SetHomeDiskSpaceProvider(fn HomeDiskSpaceFn) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.homeDiskSpaceFn = fn
}

func (d *DiskSpaceService) getUserHomeDiskSpaceInfo(ctx basecontext.ApiContext, username string) (int64, error) {
	if d.homeDiskSpaceFn == nil {
		d.ctx.LogErrorf("[DiskSpace] Home disk space provider function is not set, returning 0")
		return 0, nil
	}
	return d.homeDiskSpaceFn(ctx, username)
}

func (d *DiskSpaceService) GetDiskSpaceAvailable(ctx basecontext.ApiContext, username, folderPath string) (models.DiskSpaceAvailable, error) {
	response := models.DiskSpaceAvailable{}

	cacheDiskSpace, err := d.GetCacheDiskSpace(ctx)
	if err != nil {
		return response, err
	}
	response.CacheFolder = fmt.Sprintf("%d MB", cacheDiskSpace)

	if folderPath != "" {
		diskSpace, err := helpers.GetFreeDiskSpace(folderPath)
		if err != nil {
			return response, err
		}
		response.Given = fmt.Sprintf("%d MB", diskSpace)
	}
	parallelsHomeDiskSpace, err := d.getUserHomeDiskSpaceInfo(ctx, username)
	if err != nil {
		return response, err
	}
	response.ParallelsHome = fmt.Sprintf("%d MB", parallelsHomeDiskSpace)

	return response, nil
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

	cacheDiskSpace, err := d.GetCacheDiskSpace(d.ctx)
	if err != nil {
		d.ctx.LogErrorf("[DiskSpace] [worker] Error getting cache disk space: %v", err)
	} else {
		d.ctx.LogInfof("[DiskSpace] [worker] Cache disk space available: %d MB", cacheDiskSpace)
		event.CacheFolder = fmt.Sprintf("%d MB", cacheDiskSpace)
	}

	parallelsHomeDiskSpace, err := d.getUserHomeDiskSpaceInfo(d.ctx, "")
	if err != nil {
		d.ctx.LogErrorf("[DiskSpace] [worker] Error getting Parallels home disk space: %v", err)
	} else {
		d.ctx.LogInfof("[DiskSpace] [worker] Parallels home disk space available: %d MB", parallelsHomeDiskSpace)
		event.ParallelsHome = fmt.Sprintf("%d MB", parallelsHomeDiskSpace)
	}

	if err := ee.BroadcastMessage(models.NewEventMessage(constants.EventTypeStats, "DISK_SPACE_CHANGED", event)); err != nil {
		d.ctx.LogErrorf("[DiskSpace] [worker] Error broadcasting disk space event: %v", err)
	}
}
