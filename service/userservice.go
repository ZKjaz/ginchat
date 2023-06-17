package service

import (
	"fmt"
	"main/models"
	"main/utils"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// GetUserList
// @Summary 所有用户
// @Tags 用户模块
// @Success 200 {string} json{"code","message"}
// @Router /user/GetUserList [get]
func GetUserList(c *gin.Context) {
	data := make([]*models.UserBasic, 10)
	data = models.GetUserList()

	c.JSON(200, gin.H{
		"code":    0, // 0成功，-1失败
		"message": "获取所有用户成功！！！",
		"data":    data,
	})
}

// CreatUser
// @Summary 新增用户
// @Tags 用户模块
// @param name query string false "用户名"
// @param password query string false "密码"
// @param repassword query string false "确认密码"
// @param phone query string false "手机号"
// @param email query string false "邮箱"
// @Success 200 {string} json{"code","message"}
// @Router /user/creatUser [get]
func CreatUser(c *gin.Context) {
	user := models.UserBasic{}
	user.Name = c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	repassword := c.Request.FormValue("repassword")
	email := c.Request.FormValue("email")
	phone := c.Request.FormValue("phone")
	fmt.Println(">>>>>>>>>>>>>>", password, repassword, email, phone)

	salt := fmt.Sprintf("%06d", rand.Int31())
	data := models.FindUserByName(user.Name)
	if user.Name == "" || password == "" || repassword == "" || email == "" || phone == "" {

		c.JSON(200, gin.H{
			"code":    -1, // 0成功，-1失败
			"message": "用户名、密码、手机号或者邮箱不能为空！！！",
			"data":    user,
		})
		return
	}
	if data.Name != "" {

		c.JSON(200, gin.H{
			"code":    -1, // 0成功，-1失败
			"message": "用户名已注册！！！",
			"data":    user,
		})
		return
	}
	data = models.FindUserByPhone(phone)
	if data.Phone != "" {

		c.JSON(200, gin.H{
			"code":    -1, // 0成功，-1失败
			"message": "手机号已注册！！！",
			"data":    user,
		})
		return
	}
	data = models.FindUserByEmail(email)
	if data.Email != "" {
		c.JSON(200, gin.H{
			"code":    -1, // 0成功，-1失败
			"message": "Email已注册",
			"data":    user,
		})
		return
	}

	if password != repassword {

		c.JSON(200, gin.H{
			"code":    -1, // 0成功，-1失败
			"message": "两次密码不一致",
			"data":    user,
		})
		return
	}
	user.PassWord = utils.MakePassword(password, salt)
	user.Phone = phone
	user.Email = email
	user.Salt = salt
	models.CreatUser(user)
	c.JSON(-1, gin.H{
		"code":    0, // 0成功，-1失败
		"message": "新增用户成功！！！",
		"data":    data,
	})
}

// FindUserByNameAndPwd
// @Summary 所有用户
// @Tags 用户模块
// @param name query string false "用户名"
// @param password query string false "密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/FindUserByNameAndPwd [post]
func FindUserByNameAndPwd(c *gin.Context) {
	data := models.UserBasic{}

	// name := c.Query("name")
	// password := c.Query("password")
	name := c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	user := models.FindUserByName(name)
	if user.Name == "" {
		c.JSON(-1, gin.H{
			"code":    -1, // 0成功，-1失败
			"message": "该用户不存在！！！",
			"data":    data,
		})
		return
	}

	flag := utils.ValidPassword(password, user.Salt, user.PassWord)
	if !flag {
		c.JSON(-1, gin.H{
			"code":    -1, // 0成功，-1失败
			"message": "密码不正确！！！",
			"data":    data,
		})
		return
	}
	pwd := utils.MakePassword(password, user.Salt)
	data = models.FindUserByNameAndPwd(name, pwd)

	c.JSON(200, gin.H{
		"code":    0, // 0成功，-1失败
		"message": "登陆成功",
		"data":    data,
	})
}

// DeleteUser
// @Summary 删除用户
// @Tags 用户模块
// @param id query string false "id"
// @Success 200 {string} json{"code","message"}
// @Router /user/DeleteUser [get]
func DeleteUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.Query("id"))
	user.ID = uint(id)
	models.DeleteUser(user)

	c.JSON(200, gin.H{
		"code":    0, // 0成功，-1失败
		"message": "删除用户成功！！",
		"data":    user,
	})

}

// UpdateUser
// @Summary 修改用户
// @Tags 用户模块
// @param id formData string false "id"
// @param name formData string false "name"
// @param password formData string false "password"
// @param phone formData string false "phone"
// @param email formData string false "email"
// @Success 200 {string} json{"code","message"}
// @Router /user/UpdateUser [post]
func UpdateUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.PostForm("id"))
	user.ID = uint(id)
	user.Name = c.PostForm("name")
	user.PassWord = c.PostForm("password")
	user.Phone = c.PostForm("phone")
	user.Email = c.PostForm("email")

	//校验修改的值是否符合格式规范
	_, err := govalidator.ValidateStruct(user)
	if err != nil {
		fmt.Println(err)

		c.JSON(200, gin.H{
			"code":    -1, // 0成功，-1失败
			"message": "修改参数不正确！！！",
			"data":    user,
		})

	} else {
		models.UpdateUser(user)

		c.JSON(200, gin.H{
			"code":    0, // 0成功，-1失败
			"message": "修改用户成功！！",
			"data":    user,
		})
	}
}

// 防止跨域站点的伪造请求
var upGrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SendMsg(c *gin.Context) {
	ws, err := upGrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func(ws *websocket.Conn) {
		err = ws.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}(ws)
	MsgHandler(ws, c)
}
func MsgHandler(ws *websocket.Conn, c *gin.Context) {
	msg, err := utils.Subscribe(c, utils.PublishKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	tm := time.Now().Format("2006-01-02 15:04:05")
	m := fmt.Sprintf("[%s][%s]:", tm, msg)
	err = ws.WriteMessage(1, []byte(m))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func SendUserMsg(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}

func SerchFriends(c *gin.Context) {
	id, _ := strconv.Atoi(c.Request.FormValue("userId"))
	users := models.SerchFriend(uint(id))

	// c.JSON(200, gin.H{
	// 	"code":    0, // 0成功，-1失败
	// 	"message": "查询好友列表成功！！！",
	// 	"data":    users,
	// })
	utils.RespOKList(c.Writer, users, len(users))

}
