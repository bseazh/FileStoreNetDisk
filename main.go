package main

import (
	"FileStoreServerV1/controller"
	"fmt"
	"net/http"
)

func main() {

	// 静态资源处理
	// http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(assets.AssetFS())))
	http.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/file/upload", controller.UploadHandle)
	http.HandleFunc("/file/upload/Success", controller.UploadSucHandle)
	http.HandleFunc("/file/meta", controller.GetFileMetaHandle)
	http.HandleFunc("/file/download", controller.DownloadHandle)
	http.HandleFunc("/file/query", controller.FileQueryHandle)
	http.HandleFunc("/file/fastupload", controller.TryFastUploadHandle)

	http.HandleFunc("/user/signup", controller.SignupHandle)
	http.HandleFunc("/user/signin", controller.SignInHandle)
	http.HandleFunc("/user/info", controller.HTTPInterceptor(controller.UserInfoHandle))

	fmt.Println("开始监听(localhost:8080)端口")
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		fmt.Printf("Fail to Start Server , err : %s\n", err)
	}

}
