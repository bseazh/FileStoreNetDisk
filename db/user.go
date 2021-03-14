package db

import (
	mydb "FileStoreServerV1/db/mysql"
	"fmt"
)

type User struct {
	Username     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
}

// UserSignup : 通过用户名及密码完成user表的注册操作
func UserSignup(username string, password string) bool {
	stmt, err := mydb.DBconn().Prepare(
		"insert ignore into tbl_user (`user_name`,`user_pwd`) values (?,?)")
	if err != nil {
		fmt.Printf("Failed to Prepare , err : %s\n", err)
		return false
	}
	defer stmt.Close()

	res, err := stmt.Exec(username, password)
	if err != nil {
		fmt.Printf("Failed to insert , err : %s\n", err.Error())
		return false
	}
	if rowsAffected, err := res.RowsAffected(); err == nil && rowsAffected > 0 {
		return true
	}
	return false
}

//	UserSignin : 判断密码是否一致
func UserSignin(username string, encpwd string) bool {
	stmt, err := mydb.DBconn().Prepare("select * from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)
	if err != nil {
		fmt.Println(err.Error())
		return false
	} else if rows == nil {
		fmt.Printf("username not found : %s \n", username)
		return false
	}

	pRows := mydb.ParseRows(rows)
	if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == encpwd {
		return true
	}
	return false
}

//	UpdateToken : 刷新用户登录的token
func UpdateToken(username, token string) bool {
	stmt, err := mydb.DBconn().Prepare(
		"replace into tbl_user_token (`user_name`,`user_token`)values(?,?)")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

//	GetUserInfo : 获取用户的信息
func GetUserInfo(username string) (User, error) {
	user := User{}
	stmt, err := mydb.DBconn().Prepare(
		"select user_name,signup_at from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}
	defer stmt.Close()

	// 执行查询操作
	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		return user, err
	}
	return user, err
}
