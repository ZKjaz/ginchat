package models

import "gorm.io/gorm"

//群信息  未使用
type GroupBasic struct {
	gorm.Model
	Name    string //群名称
	OwnerId uint   //群拥有者
	Icon    string //群头像
	Desc    string //群描述
	Type    string //群类型 预留 VIP等级限制
}

func (table *GroupBasic) TableName() string {
	return "group_basic"
}
