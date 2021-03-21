package controller

import (
	dblayer "FileStoreServerV1/db"
	"FileStoreServerV1/util"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

const (
	pwd_salt = "*#890"
)

//	SignupHandle : 处理用户注册请求
func SignupHandle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		t, err := template.ParseFiles("./static/view/signup.html")
		if err != nil {
			fmt.Printf("template Parse failed , err : %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		t.Execute(w, nil)
		return
	case http.MethodPost:
		r.ParseForm()
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		if len(username) < 3 || len(password) < 3 {
			w.Write([]byte("Invaild parameter"))
			return
		}
		// 对密码进行加盐及取Sha1值加密
		enc_pwd := util.Sha1([]byte(password + pwd_salt))
		// 将用户信息注册到用户表中
		suc := dblayer.UserSignup(username, enc_pwd)
		if suc {
			w.Write([]byte("SUCCESS"))
		} else {
			w.Write([]byte("FAILED"))
		}
	}
}

//	SignInHandle : 登录接口
func SignInHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		//data, err := ioutil.ReadFile("./static/view/signin.html")
		//if err != nil {
		//	w.WriteHeader(http.StatusInternalServerError)
		//	return
		//}
		//w.Write(data)
		http.Redirect(w, r, "/static/view/signin.html", http.StatusFound)
		//w.Write([]byte("http://"+r.Host+"/static/view/signin.html"))
		return
	}

	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	encPassword := util.Sha1([]byte(password + pwd_salt))

	// 1. 校验用户名和密码
	pwdChecked := dblayer.UserSignin(username, encPassword)
	if !pwdChecked {
		w.Write([]byte("Password FAILED"))
		return
	}
	// 2. 生成访问凭证(Token)
	token := GenToken(username)
	updateRes := dblayer.UpdateToken(username, token)
	if !updateRes {
		w.Write([]byte("FAILED"))
		return
	}

	// 3. 登陆成功后重定向到首页
	//w.Write([]byte("http://"+r.Host+"/static/view/home.html"))
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	w.Write(resp.JSONBytes())
}

//	UserInfoHandle : 查询用户信息
func UserInfoHandle(w http.ResponseWriter, r *http.Request) {
	//	1 . 解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	token := r.Form.Get("token")

	//	2 . 验证token是否有效
	isVaildToken := IsTokenValid(token)
	if !isVaildToken {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	//	3 . 查询用户信息
	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	//	4.	组装并且响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	w.Write(resp.JSONBytes())

}

func GenToken(username string) string {
	// 40位字符: md5(username + timestamp + token_salt) + timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

// IsTokenValid : token是否有效
func IsTokenValid(token string) bool {
	if len(token) != 40 {
		return false
	}
	// TODO: 判断token的时效性，是否过期
	// TODO: 从数据库表tbl_user_token查询username对应的token信息
	// TODO: 对比两个token是否一致
	return true
}
