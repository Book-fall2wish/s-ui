package cronjob

import (
	"time"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/logger"
	"github.com/alireza0/s-ui/service"
)

type TrafficResetJob struct {
	service.ClientService
}

func NewTrafficResetJob() *TrafficResetJob {
	return new(TrafficResetJob)
}

func (s *TrafficResetJob) Run() {
	err := s.resetTraffic()
	if err != nil {
		logger.Warning("Reset traffic failed: ", err)
		return
	}
}

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

		// 检查是否设置了固定重置日
		if client.TrafficResetDay > 0 {
			// 如果设置了重置日，并且当前日期匹配
			if client.TrafficResetDay == currentDay {
				shouldReset = true

				// 如果设置了过期时间，检查是否在过期日期范围内
				if client.Expiry > 0 && currentTime > client.Expiry {
					// 客户端已过期，不重置流量
					shouldReset = false
				}
			}
		} else if client.LastTrafficReset > 0 {
			// 如果没有设置固定重置日，按周期重置（例如每30天）
			resetInterval := 30 * 24 * 60 * 60 // 30天的秒数
			if currentTime-client.LastTrafficReset >= resetInterval {
				shouldReset = true
			}
		}

		if shouldReset {
			logger.Debug("Resetting traffic for client: ", client.Name)

			// 重置流量使用情况
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
