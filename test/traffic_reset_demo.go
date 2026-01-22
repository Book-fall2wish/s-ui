package main

import (
	"fmt"
	"time"

	"github.com/alireza0/s-ui/database"
	"github.com/alireza0/s-ui/database/model"
	"github.com/alireza0/s-ui/service"
)

func main() {
	// 初始化数据库
	err := database.InitDB("./test.db")
	if err != nil {
		panic(err)
	}

	clientService := &service.ClientService{}

	// 创建一个测试客户端，设置每月22号重置流量
	testClient := model.Client{
		Name:             "test_user",
		Enable:           true,
		Volume:           10 * 1024 * 1024 * 1024, // 10GB
		Expiry:           time.Date(2026, 4, 22, 0, 0, 0, 0, time.UTC).Unix(), // 2026年4月22日过期
		TrafficResetDay:  22, // 每月22号重置
		LastTrafficReset: time.Now().AddDate(0, 0, -35).Unix(), // 35天前重置的，应该触发重置
		Up:               5 * 1024 * 1024 * 1024, // 已用5GB
		Down:             3 * 1024 * 1024 * 1024, // 已用3GB
	}

	db := database.GetDB()
	result := db.Create(&testClient)
	if result.Error != nil {
		fmt.Printf("创建测试客户端失败: %v\n", result.Error)
		return
	}

	fmt.Printf("测试客户端已创建，ID: %d\n", testClient.Id)

	// 调用流量重置功能
	err = clientService.ResetTrafficForClients()
	if err != nil {
		fmt.Printf("重置流量失败: %v\n", err)
		return
	}

	// 检查客户端流量是否重置
	var updatedClient model.Client
	result = db.Where("id = ?", testClient.Id).First(&updatedClient)
	if result.Error != nil {
		fmt.Printf("获取更新后的客户端失败: %v\n", result.Error)
		return
	}

	fmt.Printf("客户端流量已重置\n")
	fmt.Printf("Up: %d, Down: %d, LastTrafficReset: %d\n", updatedClient.Up, updatedClient.Down, updatedClient.LastTrafficReset)

	// 验证流量是否正确重置
	if updatedClient.Up == 0 && updatedClient.Down == 0 {
		fmt.Printf("流量重置功能测试成功！\n")
	} else {
		fmt.Printf("流量重置功能测试失败！\n")
	}
}