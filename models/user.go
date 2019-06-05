package models

import (
	"fmt"
)

type User struct {
	ID       uint32 `orm:"column(id);auto;pk" form:"-"`
	Name     string `orm:"column(name);size(64);unique" form:"name" json:"name" valid:"Required"`
	Password string `orm:"column(password);size(32)" form:"password" json:"password" valid:"Required"`
}

func Login(u *User) (ERR_CODE, error) {
	if err := o.Read(u, "name", "password"); err != nil {
		return ERR_CODE_USER_PASSWORD_NOT_MATCH, fmt.Errorf("用户名/密码错误: %+v", err)
	}

	return ERR_CODE_OK, nil
}

func ChangePassword(name, oldPassword, newPassword string) (ERR_CODE, error) {
	if name == "" || oldPassword == "" || newPassword == "" {
		return ERR_CODE_ARGS_MISS_REQUIRED, fmt.Errorf("关键参数缺失")
	}
	if oldPassword == newPassword {
		return ERR_CODE_ARGS_SAME_PASSWORD_AT_CHANGE, fmt.Errorf("新旧密码不能相同")
	}
	u := &User{
		Name:     name,
		Password: oldPassword,
	}
	if err := o.Read(u, "name", "password"); err != nil {
		return ERR_CODE_USER_PASSWORD_NOT_MATCH, fmt.Errorf("用户名/旧密码错误")
	}
	u.Password = newPassword
	if _, err := o.Update(u, "password"); err != nil {
		return ERR_CODE_UPDATE_PASSWORD_FAIL, err
	}
	return ERR_CODE_OK, nil
}