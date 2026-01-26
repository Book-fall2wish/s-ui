package cronjob

import (
	"time"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/logger"
	"github.com/alireza0/s-ui/service"
)

// TrafficResetJob handles scheduled traffic reset operations for clients.
// It supports both monthly reset on specific days and periodic 30-day cycles.
// Expired clients are automatically skipped during reset operations.
// TrafficResetJob 处理客户端的定时流量重置操作。支持每月固定日期重置和30天周期重置。过期客户端会自动跳过。
type TrafficResetJob struct {
	service.ClientService
}

// NewTrafficResetJob creates a new instance of TrafficResetJob.
// 创建新的 TrafficResetJob 实例
func NewTrafficResetJob() *TrafficResetJob {
	return new(TrafficResetJob)
}

// Run executes the traffic reset job.
// 执行流量重置任务
func (s *TrafficResetJob) Run() {
	err := s.resetTraffic()
	if err != nil {
		logger.Warning("Reset traffic failed: ", err)
		return
	}
}

// resetTraffic performs the actual traffic reset operation.
// 执行实际的流量重置操作
func (s *TrafficResetJob) resetTraffic() error {
	db := database.GetDB()
	now := time.Now()
	currentDay := now.Day()
	currentTime := now.Unix()

	var clients []model.Client
	err := db.Model(model.Client{}).Where("traffic_reset_day > 0 OR last_traffic_reset > 0").Find(&clients).Error
	if err != nil {
		return err
	}

	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	for _, client := range clients {
		shouldReset := false

		// Check if fixed reset day is set / 检查是否设置了固定重置日
		if client.TrafficResetDay > 0 {
			// If reset day is set and current date matches / 如果设置了重置日，并且当前日期匹配
			if client.TrafficResetDay == currentDay {
				shouldReset = true

				// If expiry is set, check if within expiry period / 如果设置了过期时间，检查是否在过期日期范围内
				if client.Expiry > 0 && currentTime > client.Expiry {
					// Client expired, don't reset traffic / 客户端已过期，不重置流量
					shouldReset = false
				}
			}
		} else if client.LastTrafficReset > 0 {
			// If no fixed reset day, use periodic reset (e.g., every 30 days) / 如果没有设置固定重置日，按周期重置（例如每30天）
			resetInterval := int64(30 * 24 * 60 * 60) // 30 days in seconds / 30天的秒数
			if currentTime-client.LastTrafficReset >= resetInterval {
				shouldReset = true
			}
		}

		if shouldReset {
			logger.Debug("Resetting traffic for client: ", client.Name)

			// Reset traffic usage / 重置流量使用情况
			err = tx.Model(model.Client{}).Where("id = ?", client.Id).Updates(map[string]interface{}{
				"up":                 int64(0),
				"down":               int64(0),
				"last_traffic_reset": currentTime,
			}).Error

			if err != nil {
				logger.Warning("Failed to reset traffic for client: ", client.Name, " error: ", err)
				continue
			}
		}
	}

	if err == nil {
		tx.Commit()
	} else {
		tx.Rollback()
	}

	return err
}
