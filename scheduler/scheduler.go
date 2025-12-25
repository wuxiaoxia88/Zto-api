package scheduler

import (
	"time"

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

	// 凌晨 00:05 刷新
	if hour == 0 && minute == 5 {
		logger.Token("执行凌晨定时刷新任务")
		if err := s.refreshFunc(); err != nil {
			logger.Error("凌晨定时刷新失败: %v", err)
		}
		return
	}

	// 19:30 预防性刷新
	if hour == 19 && minute == 30 {
		logger.Token("执行预防性刷新任务（避免 20:00 过期）")
		if err := s.refreshFunc(); err != nil {
			logger.Error("预防性刷新失败: %v", err)
		}
		return
	}
}

// TriggerRefresh 手动触发刷新
func (s *Scheduler) TriggerRefresh() error {
	logger.Token("手动触发 Token 刷新")
	return s.refreshFunc()
}
