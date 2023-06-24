package models

import (
	"fmt"
	"main/utils"
	"time"

	"gorm.io/gorm"
)

type UserBasic struct {
	gorm.Model
	Name       string
	PassWord   string
	Phone      string `valid:"matches(^1[3-9]{1}\\d{9}$)"`
	Email      string `valid:"email"`
	Avatar     string //头像
	Identity   string
	ClientIp   string
	ClientPort string
	Salt       string
	//LoginTime     time.Time
	//HeartbeatTime time.Time
	//LoginOutTime time.Time `gorm:"column:login_out_time" json:"login_out_time"`
	IsLogout   bool
	DeviceInfo string
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}

func GetUserList() []*UserBasic {
	date := make([]*UserBasic, 10)
	utils.DB.Find(&date)
	for _, v := range date {
		fmt.Println(v)
	}
	return date
}

func FindUserByNameAndPwd(name, password string) UserBasic {
	user := UserBasic{}

	utils.DB.Where("name = ? and pass_word = ?", name, password).First(&user)

	//token加密
	str := fmt.Sprintf("%d", time.Now().Unix())
	temp := utils.MD5Encode(str)
	utils.DB.Model(&user).Where("id = ?", user.ID).Update("Identity", temp)
	return user
}

func FindUserByName(name string) UserBasic {
	user := UserBasic{}
	utils.DB.Where("name = ?", name).First(&user)
	return user
}

func FindUserByPhone(phone string) UserBasic {
	user := UserBasic{}
	utils.DB.Where("phone = ?", phone).First(&user)
	return user
}

func FindUserByEmail(email string) UserBasic {
	user := UserBasic{}
	utils.DB.Where("email = ?", email).First(&user)
	return user
}

func CreatUser(user UserBasic) *gorm.DB {

	return utils.DB.Create(&user)
}

func DeleteUser(user UserBasic) *gorm.DB {

	return utils.DB.Delete(&user)
}

func UpdateUser(user UserBasic) *gorm.DB {

	return utils.DB.Model(&user).Updates(UserBasic{Name: user.Name, PassWord: user.PassWord, Phone: user.Phone, Email: user.Email})
}

// 查找某个用户
func FindByID(id uint) UserBasic {
	user := UserBasic{}
	utils.DB.Where("id = ?", id).First(&user)
	return user
}
