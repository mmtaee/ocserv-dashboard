package usecase

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/mmtaee/ocserv-dashboard/api/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/sync/errgroup"
)

type HomeGetHomeUser struct {
	Total  int64                      `json:"total" validate:"omitempty"`
	Online []models.OnlineUserSession `json:"online_users_session" validate:"omitempty"`
}

type HomeTelegramServiceStatus struct {
	Enabled     bool   `json:"enabled"`
	HasBotToken bool   `json:"has_bot_token"`
	BotUsername string `json:"bot_username,omitempty"`
}

type HomeGetHomeResponse struct {
	Statistics       *[]models.DailyTraffic       `json:"statistics" validate:"omitempty"`
	Users            HomeGetHomeUser              `json:"users" validate:"omitempty"`
	IPBans           *[]models.IPBanPoints        `json:"ip_bans" validate:"omitempty"`
	TopBandwidthUser repository.TopBandwidthUsers `json:"top_bandwidth_user" validate:"omitempty"`
	TotalBandwidth   repository.TotalBandwidths   `json:"total_bandwidth" validate:"omitempty"`
	TelegramService  *HomeTelegramServiceStatus   `json:"telegram_service,omitempty" validate:"omitempty"`
}

type HomeCPU struct {
	AvgPercent float64 `json:"avg_percent"`
	UsedUnits  float64 `json:"used_units"`
	Total      int     `json:"total"`
}

type HomeRAM struct {
	Used        float64 `json:"used"`
	Total       float64 `json:"total"`
	UsedPercent float64 `json:"used_percent"`
}

type HomeSwap struct {
	Used        float64 `json:"used"`
	Total       float64 `json:"total"`
	UsedPercent float64 `json:"used_percent"`
}

type HomeDisk struct {
	Used        float64 `json:"used"`
	Total       float64 `json:"total"`
	UsedPercent float64 `json:"used_percent"`
}

type HomeDockerStats struct {
	Name string  `json:"name" validate:"required"`
	CPU  HomeCPU `json:"cpu" validate:"omitempty"`
	RAM  HomeRAM `json:"ram" validate:"omitempty"`
}

type HomeDockerService struct {
	Postgres    HomeDockerStats `json:"postgres" validate:"required"`
	Ocserv      HomeDockerStats `json:"ocserv" validate:"required"`
	LogStream   HomeDockerStats `json:"log_stream" validate:"required"`
	UserExpiry  HomeDockerStats `json:"user_expiry" validate:"required"`
	TelegramBot HomeDockerStats `json:"telegram_bot" validate:"omitempty"`
	Web         HomeDockerStats `json:"web" validate:"required"`
}

type HomeServerStatusResponse struct {
	CPU  HomeCPU  `json:"cpu"`
	RAM  HomeRAM  `json:"ram"`
	Swap HomeSwap `json:"swap"`
	Disk HomeDisk `json:"disk"`
}

type HomeUsecase struct {
	occtlRepo      repository.OcctlRepositoryInterface
	ocservUserRepo repository.OcservUserRepositoryInterface
	reportRepo     repository.ReportRepositoryInterface
	telegramRepo   repository.TelegramRepositoryInterface
}

type HomeUsecaseInterface interface {
	Home() (*HomeGetHomeResponse, error)
	OcservStats() (*models.OcservStatusResponse, error)
	SystemUsageStats() (*HomeServerStatusResponse, error)
	ContainerUsageStats() (*HomeDockerService, error)
}

func NewHomeUsecase(
	occtlRepo repository.OcctlRepositoryInterface,
	ocservUserRepo repository.OcservUserRepositoryInterface,
	reportRepo repository.ReportRepositoryInterface,
	telegramRepo repository.TelegramRepositoryInterface,
) HomeUsecaseInterface {
	return &HomeUsecase{
		occtlRepo:      occtlRepo,
		ocservUserRepo: ocservUserRepo,
		reportRepo:     reportRepo,
		telegramRepo:   telegramRepo,
	}
}

func (u *HomeUsecase) Home() (*HomeGetHomeResponse, error) {
	g, ctx := errgroup.WithContext(context.Background())

	var (
		statistics       *[]models.DailyTraffic
		onlineUsers      []models.OnlineUserSession
		TotalUser        int64
		ipBans           *[]models.IPBanPoints
		topBandwidthUser repository.TopBandwidthUsers
		totalBandwidth   repository.TotalBandwidths
		telegramSnap     *HomeTelegramServiceStatus

		mu sync.Mutex
	)

	g.Go(func() error {
		data, err := u.reportRepo.TenDaysStats(ctx)
		if err != nil {
			return err
		}
		mu.Lock()
		statistics = &data
		mu.Unlock()
		return nil
	})

	g.Go(func() error {
		users, err := u.occtlRepo.OnlineSessions()
		if err != nil {
			return err
		}
		mu.Lock()
		onlineUsers = users
		mu.Unlock()
		return nil
	})

	g.Go(func() error {
		ips, err := u.occtlRepo.IPBans()
		if err != nil {
			return err
		}
		mu.Lock()
		ipBans = ips
		mu.Unlock()
		return nil
	})

	g.Go(func() error {
		users, err := u.reportRepo.TotalUsers(ctx)
		if err != nil {
			return err
		}
		mu.Lock()
		TotalUser = users
		mu.Unlock()
		return nil
	})

	g.Go(func() error {
		topUser, err := u.reportRepo.TopBandwidthUser(ctx)
		if err != nil {
			return err
		}
		mu.Lock()
		topBandwidthUser = topUser
		mu.Unlock()
		return nil
	})

	g.Go(func() error {
		bandwidth, err := u.reportRepo.TotalBandwidth(ctx)
		if err != nil {
			return err
		}
		mu.Lock()
		totalBandwidth = bandwidth
		mu.Unlock()
		return nil
	})

	g.Go(func() error {
		if os.Getenv("TELEGRAM_BOT_ENABLED") != "true" {
			telegramSnap = nil
			return nil
		}

		s, err := u.telegramRepo.Settings(ctx)
		if err != nil {
			return nil
		}
		mu.Lock()
		telegramSnap = &HomeTelegramServiceStatus{
			Enabled:     s.Enabled,
			HasBotToken: strings.TrimSpace(s.BotToken) != "",
			BotUsername: strings.TrimSpace(s.BotUsername),
		}
		mu.Unlock()
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	resp := HomeGetHomeResponse{
		Statistics: statistics,
		IPBans:     ipBans,
		Users: HomeGetHomeUser{
			Total:  TotalUser,
			Online: onlineUsers,
		},
		TopBandwidthUser: topBandwidthUser,
		TotalBandwidth:   totalBandwidth,
		TelegramService:  telegramSnap,
	}

	return &resp, nil
}

func (u *HomeUsecase) OcservStats() (*models.OcservStatusResponse, error) {
	var status models.OcservStatusResponse

	serverStatus, err := u.occtlRepo.Status()
	if err != nil {
		return nil, nil
	}
	if serverStatusMap, ok := serverStatus.(map[string]interface{}); ok {
		status = models.ParseOcservServerStatus(serverStatusMap)
	}

	return &status, nil
}

func (u *HomeUsecase) SystemUsageStats() (*HomeServerStatusResponse, error) {
	var stats HomeServerStatusResponse

	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)

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

			avg := sum / float64(len(cpuPercents))
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
		return nil, err
	}

	return &stats, nil
}

func (u *HomeUsecase) ContainerUsageStats() (*HomeDockerService, error) {
	ctx := context.Background()

	if _, err := os.Stat("/.dockerenv"); err != nil {
		return &HomeDockerService{}, nil
	}

	dockerClient, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}
	defer dockerClient.Close()

	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, err
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

	results := make(chan HomeDockerStats, len(containers))

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

			const gb = 1024 * 1024 * 1024

			usedGB := math.Round((float64(v.MemoryStats.Usage)/float64(gb))*100) / 100
			totalGB := math.Round((float64(v.MemoryStats.Limit)/float64(gb))*100) / 100

			memPercent := 0.0
			if v.MemoryStats.Limit > 0 {
				memPercent = (float64(v.MemoryStats.Usage) / float64(v.MemoryStats.Limit)) * 100
				memPercent = math.Round(memPercent*100) / 100
			}

			results <- HomeDockerStats{
				Name: name,
				CPU: HomeCPU{
					AvgPercent: avgPercent,
					UsedUnits:  usedUnits,
					Total:      totalCPUs,
				},
				RAM: HomeRAM{
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

	var service HomeDockerService

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
		return nil, err
	}

	return &service, nil
}
