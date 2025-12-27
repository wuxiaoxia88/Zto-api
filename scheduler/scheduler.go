package scheduler

import (
	"time"

	"zto-api-proxy/config"
	"zto-api-proxy/logger"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	refreshFunc func() error
	stopChan    chan struct{}
	running     bool
}

// NewScheduler 创建调度器
func NewScheduler(refreshFunc func() error) *Scheduler {
	return &Scheduler{
		refreshFunc: refreshFunc,
		stopChan:    make(chan struct{}),
	}
}

// Start 启动调度器
func (s *Scheduler) Start() {
	if s.running {
		return
	}
	s.running = true

	go s.run()
	logger.Info("定时任务调度器已启动")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	if !s.running {
		return
	}
	close(s.stopChan)
	s.running = false
	logger.Info("定时任务调度器已停止")
}

func (s *Scheduler) run() {
	// 每分钟检查一次
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case now := <-ticker.C:
			s.checkAndExecute(now)
		}
	}
}

func (s *Scheduler) checkAndExecute(now time.Time) {
	hour := now.Hour()
	minute := now.Minute()

	// 1. 凌晨 00:05 强制全面刷新 (常规维护)
	if hour == 0 && minute == 5 {
		logger.Token("执行凌晨定时刷新任务")
		if err := s.refreshFunc(); err != nil {
			logger.Error("凌晨定时刷新失败: %v", err)
		}
		return
	}

	// 2. 智能预刷新：检查是否即将失效 (提前 30 分钟)
	token := config.GetTokenData()
	if token != nil && !token.ExpiresAt.IsZero() {
		// 距离过期还有多久
		timeLeft := time.Until(token.ExpiresAt)
		if timeLeft > 0 && timeLeft < 30*time.Minute {
			logger.Token("检测到 Token 即将失效 (剩余 %v)，执行提前刷新", timeLeft.Round(time.Second))
			if err := s.refreshFunc(); err != nil {
				logger.Error("提前预刷新失败: %v", err)
			}
			return
		}
	}

	// 3. 预防型刷新 (针对 wyandyy 10h 周期，如果不幸错过了上面的检查)
	// 原 19:30 逻辑保留作为双保险
	if hour == 19 && minute == 30 {
		logger.Token("执行 19:30 预防型刷新")
		s.refreshFunc()
	}
}

// TriggerRefresh 手动触发刷新
func (s *Scheduler) TriggerRefresh() error {
	logger.Token("手动触发 Token 刷新")
	return s.refreshFunc()
}
