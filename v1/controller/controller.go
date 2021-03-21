package controller

import (
	dblayer "FileStoreServerV1/db"
	"FileStoreServerV1/meta"
	"FileStoreServerV1/util"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

// UploadHandle:	处理文件上传
func UploadHandle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// 返回 上传的页面
	case http.MethodGet:
		t, err := template.ParseFiles("./static/view/index.html")
		if err != nil {
			fmt.Printf("template Parse failed , err : %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		t.Execute(w, nil)
	// 接收文件流及存储到本地目录
	case http.MethodPost:
		//	从form获取'uploaded'属性对应的值
		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("Failed to get file data , err %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer file.Close()

		fileMeta := meta.FileMeta{
			FileName: fileHeader.Filename,
			Location: "./UploadFile/" + fileHeader.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Printf("Failed to create file, err:%s\n", err.Error())
			return
		}
		defer newFile.Close()

		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Printf("Failed to save data into file, err:%s\n", err.Error())
			return
		}

		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)
		//meta.UpdateFileMeta( fileMeta )

		_ = meta.UpdateFileMetaDB(fileMeta)

		//http.Redirect(w,r,"/file/upload/Success",http.StatusFound)
		// 更新用户文件表记录
		r.ParseForm()
		username := r.Form.Get("username")
		suc := dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1,
			fileMeta.FileName, fileMeta.FileSize)
		if suc {
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		} else {
			w.Write([]byte("Upload Failed."))
		}
	}
}

// UploadSucHandle: 上传成功跳转页面
func UploadSucHandle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("*** Upload Success! ***"))
}

// GetFileMetaHandle: 获取文件元信息
func GetFileMetaHandle(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileHash := r.Form["filehash"][0]

	//fMeta := meta.GetFileMeta( fileHash )
	fMeta, err := meta.GetFileMetaDB(fileHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(fMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// FileQueryHandle : 查询批量的文件元信息
func FileQueryHandle(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	//fileMetas, _ := meta.GetLastFileMetasDB(limitCnt)
	userFiles, err := dblayer.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// DownloadHandle: 文件下载接口
func DownloadHandle(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fsha1 := r.Form.Get("filehash")
	fm := meta.GetFileMeta(fsha1)

	f, err := os.Open(fm.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("Content-Descrption", "attachment;filename=\""+fm.FileName+"\"")
	w.Write(data)

}

//	TryFastUploadHandle : 尝试秒传接口
func TryFastUploadHandle(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// 1. 解析请求参数;
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesizeStr := r.Form.Get("filesize")
	filesize, _ := strconv.Atoi(filesizeStr)

	// 2. 从文件表中查询相同hash的文件记录;

	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Printf(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3. 查找不到记录则返回秒传失败;
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通接口上传",
		}
		w.Write(resp.JSONBytes())
		return
	}
	// 4. 上传过则将文件信息写入用户文件表,返回成功;

	suc := dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	} else {
		resp := util.RespMsg{
			Code: -2,
			Msg:  "秒传失败，请稍后重试",
		}
		w.Write(resp.JSONBytes())
		return
	}
}
