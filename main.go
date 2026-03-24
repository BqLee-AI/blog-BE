package main

import (
	"Blog/blog-BE/dao"
	"Blog/blog-BE/models"
	"Blog/blog-BE/routers"
	"log"
)

func main() {
	// 创建数据库
	if err := dao.InitPgSql(); err != nil {
		log.Fatal("数据库连接失败：", err)
	}
	// 断开数据库连接
	defer dao.MyClose()

	// 模型绑定
	dao.DB.AutoMigrate(&models.User{})

	router := routers.SetupRouter()
	// 启动服务，监听 8080 端口
	router.Run()
}
