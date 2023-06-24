package router

import (
	"main/docs"
	"main/service"

	"github.com/gin-gonic/gin"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {

	r := gin.Default()

	//swagger
	docs.SwaggerInfo.BasePath = ""
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	//静态资源
	r.Static("/asset", "asset/")
	r.StaticFile("/favicon.ico", "asset/images/favicon.ico")

	//	r.StaticFS()
	r.LoadHTMLGlob("views/**/*")

	//首页
	r.GET("/", service.GetIndex)
	r.GET("/index", service.GetIndex)
	r.GET("/toRegister", service.ToRegister)
	r.GET("/toChat", service.ToChat)
	r.GET("/chat", service.Chat)
	r.POST("/SerchFriends", service.SerchFriends)

	//用户模块
	r.GET("/user/GetUserList", service.GetUserList)
	r.POST("/user/createUser", service.CreatUser)
	r.GET("/user/DeleteUser", service.DeleteUser)
	r.POST("/user/UpdateUser", service.UpdateUser)
	r.POST("/user/findUserByNameAndPwd", service.FindUserByNameAndPwd)
	r.POST("/user/find", service.FindByID)

	//发送消息
	r.GET("/user/SendMsg", service.SendMsg)
	//发送消息
	r.GET("/user/SendUserMsg", service.SendUserMsg)
	//上传
	r.POST("/attach/upload", service.Upload)
	//添加好友
	r.POST("/contact/addfriend", service.AddFriend)
	//创建群
	r.POST("/contact/createCommunity", service.CreateCommunity)
	//群列表
	r.POST("/contact/loadcommunity", service.LoadCommunity)
	//加入群聊

	r.POST("/contact/joinGroup", service.JoinGroup)
	r.POST("/user/redisMsg", service.RedisMsg)

	return r

}
