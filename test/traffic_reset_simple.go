package main

import (
	"fmt"
	"time"
)

// Simplified Client structure for testing logic only
// 简化的 Client 结构，仅用于测试逻辑
type Client struct {
	Id               int64
	Name             string
	Volume           int64
	Expiry           int64
	Up               int64
	Down             int64
	TrafficResetDay  int
	LastTrafficReset int64
}

// shouldResetTraffic determines if a client's traffic should be reset based on configuration
// shouldResetTraffic 根据配置判断是否应该重置客户端流量
func shouldResetTraffic(client Client, now time.Time) bool {
	currentDay := now.Day()
	currentTime := now.Unix()

	// Check if fixed reset day is set / 检查是否设置了固定重置日
	if client.TrafficResetDay > 0 {
		// If reset day is set and current date matches / 如果设置了重置日，并且当前日期匹配
		if client.TrafficResetDay == currentDay {
			shouldReset := true
			
			// If expiry is set, check if within expiry period / 如果设置了过期时间，检查是否在过期日期范围内
			if client.Expiry > 0 && currentTime > client.Expiry {
				// Client expired, don't reset traffic / 客户端已过期，不重置流量
				shouldReset = false
			}
			return shouldReset
		}
	} else if client.LastTrafficReset > 0 {
		// If no fixed reset day, use periodic reset (e.g., every 30 days) / 如果没有设置固定重置日，按周期重置（例如每30天）
		resetInterval := int64(30 * 24 * 60 * 60) // 30 days in seconds / 30天的秒数
		if currentTime-client.LastTrafficReset >= resetInterval {
			return true
		}
	}

	return false
}

func main() {
	now := time.Now()
	fmt.Printf("Current time: %s (Day: %d) / 当前时间: %s (日期: %d)\n", now.Format("2006-01-02 15:04:05"), now.Day())
	fmt.Println()

	// Test case 1: Reset on the 22nd of each month, today is the 22nd, should reset
	// 测试用例1: 每月22号重置，今天是22号，应该重置
	client1 := Client{
		Id:               1,
		Name:             "Monthly reset user (22nd) / 每月22号重置用户",
		TrafficResetDay:  22,
		Expiry:           time.Now().AddDate(0, 2, 0).Unix(), // Expires in 2 months / 2个月后过期
		Up:               5 * 1024 * 1024 * 1024,
		Down:             3 * 1024 * 1024 * 1024,
	}
	result1 := shouldResetTraffic(client1, now)
	fmt.Printf("Test 1 - %s: %v (Expected: true / 预期: true)\n", client1.Name, result1)

	// Test case 2: Reset on the 1st of each month, today is the 22nd, should not reset
	// 测试用例2: 每月1号重置，今天是22号，不应该重置
	client2 := Client{
		Id:               2,
		Name:             "Monthly reset user (1st) / 每月1号重置用户",
		TrafficResetDay:  1,
		Up:               5 * 1024 * 1024 * 1024,
		Down:             3 * 1024 * 1024 * 1024,
	}
	result2 := shouldResetTraffic(client2, now)
	fmt.Printf("Test 2 - %s: %v (Expected: false / 预期: false)\n", client2.Name, result2)

	// Test case 3: Reset on the 22nd of each month, but expired, should not reset
	// 测试用例3: 每月22号重置，但已过期，不应该重置
	client3 := Client{
		Id:               3,
		Name:             "Expired user / 已过期用户",
		TrafficResetDay:  22,
		Expiry:           time.Now().AddDate(0, -1, 0).Unix(), // Expired 1 month ago / 1个月前已过期
		Up:               5 * 1024 * 1024 * 1024,
		Down:             3 * 1024 * 1024 * 1024,
	}
	result3 := shouldResetTraffic(client3, now)
	fmt.Printf("Test 3 - %s: %v (Expected: false / 预期: false)\n", client3.Name, result3)

	// Test case 4: Periodic reset (every 30 days), last reset was 35 days ago, should reset
	// 测试用例4: 周期性重置（每30天），上次重置是35天前，应该重置
	client4 := Client{
		Id:               4,
		Name:             "Periodic reset user / 周期重置用户",
		LastTrafficReset: time.Now().AddDate(0, 0, -35).Unix(), // 35 days ago / 35天前
		Up:               5 * 1024 * 1024 * 1024,
		Down:             3 * 1024 * 1024 * 1024,
	}
	result4 := shouldResetTraffic(client4, now)
	fmt.Printf("Test 4 - %s: %v (Expected: true / 预期: true)\n", client4.Name, result4)

	// Test case 5: Periodic reset (every 30 days), last reset was 20 days ago, should not reset
	// 测试用例5: 周期性重置（每30天），上次重置是20天前，不应该重置
	client5 := Client{
		Id:               5,
		Name:             "Periodic reset user (not due) / 周期重置用户(未到期)",
		LastTrafficReset: time.Now().AddDate(0, 0, -20).Unix(), // 20 days ago / 20天前
		Up:               5 * 1024 * 1024 * 1024,
		Down:             3 * 1024 * 1024 * 1024,
	}
	result5 := shouldResetTraffic(client5, now)
	fmt.Printf("Test 5 - %s: %v (Expected: false / 预期: false)\n", client5.Name, result5)

	// Test case 6: No reset rules set, should not reset
	// 测试用例6: 没有设置任何重置规则，不应该重置
	client6 := Client{
		Id:       6,
		Name:     "User without reset rules / 无重置规则用户",
		Up:       5 * 1024 * 1024 * 1024,
		Down:     3 * 1024 * 1024 * 1024,
	}
	result6 := shouldResetTraffic(client6, now)
	fmt.Printf("Test 6 - %s: %v (Expected: false / 预期: false)\n", client6.Name, result6)

	fmt.Println()
	
	// Count test results / 统计测试结果
	totalTests := 6
	passedTests := 0
	if result1 { passedTests++ }
	if !result2 { passedTests++ }
	if !result3 { passedTests++ }
	if result4 { passedTests++ }
	if !result5 { passedTests++ }
	if !result6 { passedTests++ }

	fmt.Printf("Test results: %d/%d passed / 测试结果: %d/%d 通过\n", passedTests, totalTests, passedTests, totalTests)
	if passedTests == totalTests {
		fmt.Println("✅ All tests passed! Traffic reset logic is working correctly. / 所有测试通过！流量重置逻辑正常工作。")
	} else {
		fmt.Println("❌ Some tests failed! Please check the logic. / 部分测试失败！请检查逻辑。")
	}
}
