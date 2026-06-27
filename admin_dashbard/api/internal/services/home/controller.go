package home

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/api/pkg/request"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"
	"golang.org/x/sync/errgroup"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	"net/http"
	"sync"

	_ "github.com/docker/docker/api/types"
)

type Controller struct {
	request        request.CustomRequestInterface
	occtlRepo      repository.OcctlRepositoryInterface
	ocservUserRepo repository.OcservUserRepositoryInterface
	reportRepo     repository.ReportRepositoryInterface
	telegramRepo   repository.TelegramRepositoryInterface
}

func New() *Controller {
	return &Controller{
		request:        request.NewCustomRequest(),
		occtlRepo:      repository.NewOcctlRepository(),
		ocservUserRepo: repository.NewtOcservUserRepository(),
		reportRepo:     repository.NewtReportRepository(),
		telegramRepo:   repository.NewTelegramRepository(),
	}
}

// Home 	     Content of home
//
// @Summary      Content of home
// @Description  Content of home
// @Tags         Home
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object} GetHomeResponse
// @Router       /home [get]
func (ctl *Controller) Home(c echo.Context) error {
	ctx := c.Request().Context()
	g, ctx := errgroup.WithContext(ctx)

	var (
		statistics       *[]models.DailyTraffic
		onlineUsers      []models.OnlineUserSession
		TotalUser        int64
		ipBans           *[]models.IPBanPoints
		topBandwidthUser repository.TopBandwidthUsers
		totalBandwidth   repository.TotalBandwidths
		telegramSnap     *TelegramServiceStatus

		mu sync.Mutex
	)

	// -----------------------------
	// 10 days stats
	g.Go(func() error {
		data, err := ctl.reportRepo.TenDaysStats(ctx)
		if err != nil {
			return err
		}
		mu.Lock()
		statistics = &data
		mu.Unlock()
		return nil
	})

	// -----------------------------
	// online users
	g.Go(func() error {
		users, err := ctl.occtlRepo.OnlineSessions()
		if err != nil {
			return err
		}
		mu.Lock()
		onlineUsers = users
		mu.Unlock()
		return nil
	})

	// -----------------------------
	// IP bans
	g.Go(func() error {
		ips, err := ctl.occtlRepo.IPBans()
		if err != nil {
			return err
		}
		mu.Lock()
		ipBans = ips
		mu.Unlock()
		return nil
	})

	// -----------------------------
	// total users
	g.Go(func() error {
		users, err := ctl.reportRepo.TotalUsers(ctx)
		if err != nil {
			return err
		}
		mu.Lock()
		TotalUser = users
		mu.Unlock()
		return nil
	})

	// -----------------------------
	// top bandwidth user
	g.Go(func() error {
		topUser, err := ctl.reportRepo.TopBandwidthUser(ctx)
		if err != nil {
			return err
		}
		mu.Lock()
		topBandwidthUser = topUser
		mu.Unlock()
		return nil
	})

	// -----------------------------
	// total bandwidth
	g.Go(func() error {
		bandwidth, err := ctl.reportRepo.TotalBandwidth(ctx)
		if err != nil {
			return err
		}
		mu.Lock()
		totalBandwidth = bandwidth
		mu.Unlock()
		return nil
	})

	// telegram settings snapshot (no external API calls)
	g.Go(func() error {
		if os.Getenv("TELEGRAM_BOT_ENABLED") != "true" {
			telegramSnap = nil
			return nil
		}

		s, err := ctl.telegramRepo.Settings(ctx)
		if err != nil {
			return nil
		}
		mu.Lock()
		telegramSnap = &TelegramServiceStatus{
			Enabled:     s.Enabled,
			HasBotToken: strings.TrimSpace(s.BotToken) != "",
			BotUsername: strings.TrimSpace(s.BotUsername),
		}
		mu.Unlock()
		return nil
	})

	// -----------------------------
	// WAIT ALL (IMPORTANT)
	if err := g.Wait(); err != nil {
		logger.Warn("error in Home handler: %v", err)
		return ctl.request.BadRequest(c, err)
	}

	resp := GetHomeResponse{
		Statistics: statistics,
		IPBans:     ipBans,
		Users: GetHomeUser{
			Total:  TotalUser,
			Online: onlineUsers,
		},
		TopBandwidthUser: topBandwidthUser,
		TotalBandwidth:   totalBandwidth,
		TelegramService:  telegramSnap,
	}

	return c.JSON(http.StatusOK, resp)
}

// OcservStats 	     Content of ocserv server stats
//
// @Summary      Content of ocserv server stats
// @Description  Content of ocserv server stats
// @Tags         Home
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object} OcservStatusResponse
// @Router       /home/ocserv-stats [get]
func (ctl *Controller) OcservStats(c echo.Context) error {
	var status OcservStatusResponse

	serverStatus, err := ctl.occtlRepo.Status()
	if err != nil {
		return nil
	}
	if serverStatusMap, ok := serverStatus.(map[string]interface{}); ok {
		status = ParseServerStatus(serverStatusMap)
	}

	return c.JSON(http.StatusOK, status)
}

// SystemUsageStats Content of os system usage stats
//
// @Summary      Content of os system usage stats
// @Description  Content of os system usage stats (cpu, ram, swap)
// @Tags         Home
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object} ServerStatusResponse
// @Router       /home/system-stats [get]
func (ctl *Controller) SystemUsageStats(c echo.Context) error {
	var stats ServerStatusResponse

	ctx := c.Request().Context()
	g, ctx := errgroup.WithContext(ctx)

	// -----------------------------
	// CPU
	g.Go(func() error {
		cpuPercents, err := cpu.Percent(time.Second, true)
		if err != nil {
			return err
		}

		if len(cpuPercents) > 0 {
			var sum float64

			for _, p := range cpuPercents {
				sum += p
			}

			avg := sum / float64(len(cpuPercents)) // average %

			cpuCount := len(cpuPercents)

			usedUnits := (avg / 100) * float64(cpuCount)

			stats.CPU.AvgPercent = math.Round(avg*100) / 100
			stats.CPU.UsedUnits = math.Round(usedUnits*100) / 100
		}

		cpuTotal, err := cpu.Counts(true)
		if err != nil {
			return err
		}
		stats.CPU.Total = cpuTotal

		return nil
	})

	// -----------------------------
	// RAM
	g.Go(func() error {
		vm, err := mem.VirtualMemory()
		if err != nil {
			return err
		}

		const gb = 1024 * 1024 * 1024

		stats.RAM.Used = math.Round((float64(vm.Used)/float64(gb))*100) / 100
		stats.RAM.Total = math.Round((float64(vm.Total)/float64(gb))*100) / 100
		stats.RAM.UsedPercent = math.Round(vm.UsedPercent*100) / 100

		return nil
	})

	// -----------------------------
	// SWAP
	g.Go(func() error {
		sw, err := mem.SwapMemory()
		if err != nil {
			return err
		}

		const gb = 1024 * 1024 * 1024

		stats.Swap.Used = math.Round((float64(sw.Used)/float64(gb))*100) / 100
		stats.Swap.Total = math.Round((float64(sw.Total)/float64(gb))*100) / 100
		stats.Swap.UsedPercent = math.Round(sw.UsedPercent*100) / 100

		return nil
	})

	g.Go(func() error {
		usage, err := disk.Usage("/")
		if err != nil {
			return err
		}

		const gb = 1024 * 1024 * 1024

		stats.Disk.Used = math.Round((float64(usage.Used)/float64(gb))*100) / 100
		stats.Disk.Total = math.Round((float64(usage.Total)/float64(gb))*100) / 100
		stats.Disk.UsedPercent = math.Round(usage.UsedPercent*100) / 100

		return nil
	})

	if err := g.Wait(); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, stats)
}

// ContainerUsageStats Content of docker system usage stats
//
// @Summary      Content of docker system usage stats
// @Description  Content of docker system usage stats (cpu, ram, swap)
// @Tags         Home
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure      400 {object} request.ErrorResponse
// @Failure      401 {object} middlewares.Unauthorized
// @Success      200  {object} []DockerService
// @Router       /home/container-stats [get]
func (ctl *Controller) ContainerUsageStats(c echo.Context) error {
	ctx := c.Request().Context()

	if _, err := os.Stat("/.dockerenv"); err != nil {
		return c.JSON(http.StatusOK, DockerService{})
	}

	dockerClient, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}
	defer dockerClient.Close()

	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return ctl.request.BadRequest(c, err)
	}

	target := map[string]bool{
		"ocserv":          true,
		"log_stream":      true,
		"user_expiry":     true,
		"web":             true,
		"ocserv-postgres": true,
	}

	if os.Getenv("TELEGRAM_BOT_ENABLED") == "true" {
		target["telegram_bot"] = true
	}

	results := make(chan DockerStats, len(containers))

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(5)

	for _, ctr := range containers {
		ctr := ctr

		if len(ctr.Names) == 0 {
			continue
		}

		name := strings.TrimPrefix(ctr.Names[0], "/")
		if !target[name] {
			continue
		}

		g.Go(func() error {
			stat, err := dockerClient.ContainerStats(gctx, ctr.ID, false)
			if err != nil {
				return nil
			}
			defer stat.Body.Close()

			var v container.StatsResponse
			if err := json.NewDecoder(stat.Body).Decode(&v); err != nil {
				return nil
			}

			// ===== CPU =====
			cpuDelta := float64(v.CPUStats.CPUUsage.TotalUsage - v.PreCPUStats.CPUUsage.TotalUsage)
			systemDelta := float64(v.CPUStats.SystemUsage - v.PreCPUStats.SystemUsage)

			avgPercent := 0.0
			totalCPUs := int(v.CPUStats.OnlineCPUs)
			if totalCPUs == 0 {
				totalCPUs = len(v.CPUStats.CPUUsage.PercpuUsage)
			}

			if cpuDelta > 0 && systemDelta > 0 && totalCPUs > 0 {
				avgPercent = (cpuDelta / systemDelta) * float64(totalCPUs) * 100
				avgPercent = math.Round(avgPercent*100) / 100
			}

			usedUnits := math.Round(((avgPercent/100)*float64(totalCPUs))*100) / 100

			// ===== RAM =====
			const gb = 1024 * 1024 * 1024

			usedGB := math.Round((float64(v.MemoryStats.Usage)/float64(gb))*100) / 100
			totalGB := math.Round((float64(v.MemoryStats.Limit)/float64(gb))*100) / 100

			memPercent := 0.0
			if v.MemoryStats.Limit > 0 {
				memPercent = (float64(v.MemoryStats.Usage) / float64(v.MemoryStats.Limit)) * 100
				memPercent = math.Round(memPercent*100) / 100
			}

			results <- DockerStats{
				Name: name,
				CPU: CPU{
					AvgPercent: avgPercent,
					UsedUnits:  usedUnits,
					Total:      totalCPUs,
				},
				RAM: RAM{
					Used:        usedGB,
					Total:       totalGB,
					UsedPercent: memPercent,
				},
			}

			return nil
		})
	}

	go func() {
		_ = g.Wait()
		close(results)
	}()

	// ✅ Build final struct
	var service DockerService

	for r := range results {
		switch r.Name {
		case "ocserv-postgres":
			service.Postgres = r
		case "ocserv":
			service.Ocserv = r
		case "log_stream":
			service.LogStream = r
		case "user_expiry":
			service.UserExpiry = r
		case "telegram_bot":
			service.TelegramBot = r
		case "web":
			service.Web = r
		}
	}

	if err := g.Wait(); err != nil {
		return ctl.request.BadRequest(c, err)
	}

	return c.JSON(http.StatusOK, service)
}
