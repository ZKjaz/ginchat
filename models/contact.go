package models

import (
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

// 搜索好友关系
func SerchFriend(userId uint) []UserBasic {
	contacts := make([]Contact, 0)
	objIds := make([]uint64, 0)
	utils.DB.Where("owner_id = ? and type=1", userId).Find(&contacts)
	for _, v := range contacts {
		//	fmt.Println(">>>>>>>>>>>>>>", v)
		objIds = append(objIds, uint64(v.TargetId))
	}
	users := make([]UserBasic, 0)
	utils.DB.Where("id in ?", objIds).Find(&users)
	return users
}

// 添加好友   自己的ID  ， 好友的ID
func AddFriend(userId uint, targetId uint) (int, string) {
	//	user := UserBasic{}
	if targetId != 0 {
		targetUser := FindByID(targetId)
		if targetUser.Name != "" {
			if targetUser.ID == userId {
				return -1, "不能添加自己为好友！！！"
			}

			contact0 := Contact{}
			utils.DB.Where("owner_id = ? and target_id = ? and type =1", userId, targetUser.ID).Find(&contact0)
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
			contact.TargetId = targetUser.ID
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
	return -1, "好友ID不能为空!!!!"
}

// 加入群组
func JoinGroup(id uint, comId string) (int, string) {
	contact := Contact{}
	contact.OwnerId = id
	contact.Type = 2

	community := Community{}

	utils.DB.Where("id = ? or name = ?", comId, comId).Find(&community)
	if community.Name == "" {
		return -1, "没有找到群组！！！"
	}
	utils.DB.Where("owner_id = ? and target_id = ? and type = 2", id, comId).Find(&contact)
	if !contact.CreatedAt.IsZero() {
		return -1, "已经加过此群！！！"
	} else {
		contact.TargetId = community.ID
		utils.DB.Create(&contact)
		return 0, "加群成功！！！"
	}

}

func SearchUserByGroupId(communityId uint) []uint {
	contacts := make([]Contact, 0)
	objIds := make([]uint, 0)
	utils.DB.Where("target_id = ? and type=2", communityId).Find(&contacts)
	for _, v := range contacts {
		objIds = append(objIds, uint(v.OwnerId))
	}
	return objIds
}
