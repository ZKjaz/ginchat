package models

import (
	"fmt"
	"main/utils"

	"gorm.io/gorm"
)

// 群表
type Community struct {
	gorm.Model
	Name    string //群名称
	OwnerId uint   //创建者的ID
	Img     string //群头像
	Desc    string //群备注
}

func CreateCommunity(community Community) (int, string) {
	tx := utils.DB.Begin()
	//事务一旦开始，不论什么异常最终都会 Rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if len(community.Name) < 3 {
		return -1, "群名称长度小于3"
	}
	if community.OwnerId == 0 {
		return -1, "不存在此群,尝试重新登陆"
	}

	//尝试在community建群
	if err := utils.DB.Create(&community).Error; err != nil {
		fmt.Println(err)
		tx.Rollback()
		return -1, "建群失败"
	}

	//在Contact建立群关系
	contact := Contact{}
	contact.OwnerId = community.OwnerId
	contact.TargetId = community.ID
	contact.Type = 2 //群关系
	if err := utils.DB.Create(&contact).Error; err != nil {
		fmt.Println(err)
		tx.Rollback()
		return -1, "建群失败！！！"
	}
	tx.Commit()
	return 0, "建群成功！！！"
}

func LoadCommunity(ownerId uint) ([]*Community, string) {

	//我加入的群有哪些
	contact := make([]Contact, 0)
	objIds := make([]uint64, 0)
	utils.DB.Where("type = 2 and owner_id = ?", ownerId).Find(&contact)
	for _, v := range contact {
		objIds = append(objIds, uint64(v.TargetId))
	}

	//我创建的群有哪些
	data := make([]*Community, 10)
	utils.DB.Where("id in ? ", objIds).Find(&data)
	for _, v := range data {
		fmt.Println(v)
	}

	return data, "加载成功"
}
