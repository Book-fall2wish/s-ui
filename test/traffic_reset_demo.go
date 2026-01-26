package main

import (
	"fmt"
	"time"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/service"
)

func main() {
	// Initialize database / 初始化数据库
	err := database.InitDB("./test.db")
	if err != nil {
		panic(err)
	}

	clientService := &service.ClientService{}

	// Create a test client with monthly traffic reset on the 22nd / 创建一个测试客户端，设置每月22号重置流量
	testClient := model.Client{
		Name:             "test_user",
		Enable:           true,
		Volume:           10 * 1024 * 1024 * 1024, // 10GB
		Expiry:           time.Date(2026, 4, 22, 0, 0, 0, 0, time.UTC).Unix(), // Expires on April 22, 2026 / 2026年4月22日过期
		TrafficResetDay:  22, // Reset on the 22nd of each month / 每月22号重置
		LastTrafficReset: time.Now().AddDate(0, 0, -35).Unix(), // Reset 35 days ago, should trigger reset / 35天前重置的，应该触发重置
		Up:               5 * 1024 * 1024 * 1024, // Used 5GB / 已用5GB
		Down:             3 * 1024 * 1024 * 1024, // Used 3GB / 已用3GB
	}

	db := database.GetDB()
	result := db.Create(&testClient)
	if result.Error != nil {
		fmt.Printf("Failed to create test client: %v / 创建测试客户端失败: %v\n", result.Error)
		return
	}

	fmt.Printf("Test client created, ID: %d / 测试客户端已创建，ID: %d\n", testClient.Id)

	// Call traffic reset function / 调用流量重置功能
	err = clientService.ResetTrafficForClients()
	if err != nil {
		fmt.Printf("Failed to reset traffic: %v / 重置流量失败: %v\n", err)
		return
	}

	// Check if client traffic has been reset / 检查客户端流量是否重置
	var updatedClient model.Client
	result = db.Where("id = ?", testClient.Id).First(&updatedClient)
	if result.Error != nil {
		fmt.Printf("Failed to get updated client: %v / 获取更新后的客户端失败: %v\n", result.Error)
		return
	}

	fmt.Printf("Client traffic has been reset / 客户端流量已重置\n")
	fmt.Printf("Up: %d, Down: %d, LastTrafficReset: %d\n", updatedClient.Up, updatedClient.Down, updatedClient.LastTrafficReset)

	// Verify if traffic is correctly reset / 验证流量是否正确重置
	if updatedClient.Up == 0 && updatedClient.Down == 0 {
		fmt.Printf("Traffic reset function test passed! / 流量重置功能测试成功！\n")
	} else {
		fmt.Printf("Traffic reset function test failed! / 流量重置功能测试失败！\n")
	}
}