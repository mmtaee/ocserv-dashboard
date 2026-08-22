package dashboard

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/models"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	"golang.org/x/sync/errgroup"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type OCCTL interface {
	OnlineSessions() ([]models.OnlineUserSession, error)
	IPBans() (*[]models.IPBanPoints, error)
	Status() (interface{}, error)
}

type Reports interface {
	TenDaysStats(ctx context.Context) ([]models.DailyTraffic, error)
	TotalUsers(ctx context.Context) (int64, error)
	TopBandwidthUser(ctx context.Context) (repository.TopBandwidthUsers, error)
	TotalBandwidth(ctx context.Context) (repository.TotalBandwidths, error)
}

type Telegram interface {
	Settings(ctx context.Context) (*models.TelegramSettings, error)
}

type Usecase struct {
	occtl           OCCTL
	reports         Reports
	telegram        Telegram
	telegramEnabled bool
}

func New(occtl OCCTL, reports Reports, telegram Telegram, telegramEnabled bool) *Usecase {
	return &Usecase{occtl: occtl, reports: reports, telegram: telegram, telegramEnabled: telegramEnabled}
}

func (u *Usecase) Home(ctx context.Context) (*GetHomeResponse, error) {
	group, groupCtx := errgroup.WithContext(ctx)
	var statistics *[]models.DailyTraffic
	var online []models.OnlineUserSession
	var totalUsers int64
	var bans *[]models.IPBanPoints
	var top repository.TopBandwidthUsers
	var bandwidth repository.TotalBandwidths
	var telegramStatus *TelegramServiceStatus

	group.Go(func() error { value, err := u.reports.TenDaysStats(groupCtx); statistics = &value; return err })
	group.Go(func() error { value, err := u.occtl.OnlineSessions(); online = value; return err })
	group.Go(func() error { value, err := u.occtl.IPBans(); bans = value; return err })
	group.Go(func() error { value, err := u.reports.TotalUsers(groupCtx); totalUsers = value; return err })
	group.Go(func() error { value, err := u.reports.TopBandwidthUser(groupCtx); top = value; return err })
	group.Go(func() error { value, err := u.reports.TotalBandwidth(groupCtx); bandwidth = value; return err })
	if u.telegramEnabled {
		group.Go(func() error {
			settings, err := u.telegram.Settings(groupCtx)
			if err != nil {
				return nil
			}
			telegramStatus = &TelegramServiceStatus{
				Enabled: settings.Enabled, HasBotToken: strings.TrimSpace(settings.BotToken) != "", BotUsername: strings.TrimSpace(settings.BotUsername),
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	return &GetHomeResponse{
		Statistics: statistics, IPBans: bans, Users: GetHomeUser{Total: totalUsers, Online: online},
		TopBandwidthUser: top, TotalBandwidth: bandwidth, TelegramService: telegramStatus,
	}, nil
}

func (u *Usecase) OcservStats() OcservStatusResponse {
	status, err := u.occtl.Status()
	if err != nil {
		return OcservStatusResponse{}
	}
	values, ok := status.(map[string]interface{})
	if !ok {
		return OcservStatusResponse{}
	}
	return ParseServerStatus(values)
}

func (u *Usecase) SystemUsage(ctx context.Context) (*ServerStatusResponse, error) {
	group, _ := errgroup.WithContext(ctx)
	var stats ServerStatusResponse
	group.Go(func() error {
		values, err := cpu.Percent(time.Second, true)
		if err != nil {
			return err
		}
		if len(values) > 0 {
			var sum float64
			for _, value := range values {
				sum += value
			}
			average := sum / float64(len(values))
			stats.CPU.AvgPercent = round(average)
			stats.CPU.UsedUnits = round((average / 100) * float64(len(values)))
		}
		total, err := cpu.Counts(true)
		stats.CPU.Total = total
		return err
	})
	group.Go(func() error {
		value, err := mem.VirtualMemory()
		if err == nil {
			stats.RAM = RAM{Used: bytesToGB(value.Used), Total: bytesToGB(value.Total), UsedPercent: round(value.UsedPercent)}
		}
		return err
	})
	group.Go(func() error {
		value, err := mem.SwapMemory()
		if err == nil {
			stats.Swap = Swap{Used: bytesToGB(value.Used), Total: bytesToGB(value.Total), UsedPercent: round(value.UsedPercent)}
		}
		return err
	})
	group.Go(func() error {
		value, err := disk.Usage("/")
		if err == nil {
			stats.Disk = Disk{Used: bytesToGB(value.Used), Total: bytesToGB(value.Total), UsedPercent: round(value.UsedPercent)}
		}
		return err
	})
	if err := group.Wait(); err != nil {
		return nil, err
	}
	return &stats, nil
}

func (u *Usecase) ContainerUsage(ctx context.Context) (*DockerService, error) {
	if _, err := os.Stat("/.dockerenv"); err != nil {
		return &DockerService{}, nil
	}
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	defer dockerClient.Close()
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, err
	}
	targets := map[string]bool{"ocserv": true, "backend": true, "web": true, "ocserv-postgres": true}
	results := make(chan DockerStats, len(containers))
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(5)
	for _, item := range containers {
		item := item
		if len(item.Names) == 0 || !targets[strings.TrimPrefix(item.Names[0], "/")] {
			continue
		}
		group.Go(func() error {
			stats, err := dockerClient.ContainerStats(groupCtx, item.ID, false)
			if err != nil {
				return nil
			}
			defer stats.Body.Close()
			var value container.StatsResponse
			if err := json.NewDecoder(stats.Body).Decode(&value); err != nil {
				return nil
			}
			results <- containerStats(strings.TrimPrefix(item.Names[0], "/"), value)
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	close(results)
	var services DockerService
	for result := range results {
		switch result.Name {
		case "ocserv-postgres":
			services.Postgres = result
		case "ocserv":
			services.Ocserv = result
		case "backend":
			services.Backend = result
		case "web":
			services.Web = result
		}
	}
	return &services, nil
}

func containerStats(name string, value container.StatsResponse) DockerStats {
	totalCPUs := int(value.CPUStats.OnlineCPUs)
	if totalCPUs == 0 {
		totalCPUs = len(value.CPUStats.CPUUsage.PercpuUsage)
	}
	cpuDelta := float64(value.CPUStats.CPUUsage.TotalUsage - value.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(value.CPUStats.SystemUsage - value.PreCPUStats.SystemUsage)
	var average float64
	if cpuDelta > 0 && systemDelta > 0 && totalCPUs > 0 {
		average = round((cpuDelta / systemDelta) * float64(totalCPUs) * 100)
	}
	var memoryPercent float64
	if value.MemoryStats.Limit > 0 {
		memoryPercent = round(float64(value.MemoryStats.Usage) / float64(value.MemoryStats.Limit) * 100)
	}
	return DockerStats{Name: name,
		CPU: CPU{AvgPercent: average, UsedUnits: round((average / 100) * float64(totalCPUs)), Total: totalCPUs},
		RAM: RAM{Used: bytesToGB(value.MemoryStats.Usage), Total: bytesToGB(value.MemoryStats.Limit), UsedPercent: memoryPercent},
	}
}

func round(value float64) float64    { return math.Round(value*100) / 100 }
func bytesToGB(value uint64) float64 { return round(float64(value) / (1024 * 1024 * 1024)) }
