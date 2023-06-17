package models

import (
	"fmt"
	"main/utils"

	"gorm.io/gorm"
)

// 人员关系
type Contact struct {
	gorm.Model
	OwnerId  uint //谁的关系信息
	TargetId uint //对应的关系
	Type     int  //对应的类型  1好友 2群组
	Desc     string
}

func (table *Contact) TableName() string {
	return "contact"
}

func SerchFriend(userId uint) []UserBasic {
	contacts := make([]Contact, 0)
	objIds := make([]uint64, 0)
	utils.DB.Where("owner_id = ? and type=1", userId).Find(&contacts)
	for _, v := range contacts {
		fmt.Println(">>>>>>>>>>>>>>", v)
		objIds = append(objIds, uint64(v.TargetId))
	}
	users := make([]UserBasic, 0)
	utils.DB.Where("id in ?", objIds).Find(&users)
	return users
}

func AddFriend(userId uint, targetId uint) (int, string) {
	user := UserBasic{}
	if targetId != 0 {
		user = FindByID(targetId)
		if user.Name != "" {
			if targetId == userId {
				return -1, "不能添加自己为好友！！！"
			}

			contact0 := Contact{}
			utils.DB.Where("owner_id = ? and target_id = ? and type =1", userId, targetId).Find(&contact0)
			if contact0.ID != 0 {
				return -1, "该用户已经被添加！！！"
			}

			tx := utils.DB.Begin()

			//事务一旦开始，不论什么异常最终都会Rollback
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
				}
			}()

			contact := Contact{}
			contact.OwnerId = userId
			contact.TargetId = targetId
			contact.Type = 1
			if err := utils.DB.Create(&contact).Error; err != nil {
				tx.Rollback()
				return -1, "目标用户查找失败！！！！"
			}

			contact1 := Contact{}
			contact1.OwnerId = targetId
			contact1.TargetId = userId
			contact1.Type = 1
			if err := utils.DB.Create(&contact1).Error; err != nil {
				tx.Rollback()
				return -1, "目标用户查找失败！！！！"
			}
			tx.Commit()
			return 0, "添加好友关系成功！！！"
		}

	}
	return -1, "好友ID不能为空！！！！"
}
