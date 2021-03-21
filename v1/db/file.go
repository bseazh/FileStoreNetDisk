package db

import (
	mydb "FileStoreServerV1/db/mysql"
	"database/sql"
	"fmt"
)

//	OnFileUploadFinished:文件上传完成,保存meta
func OnFileUploadFinished(fileHash string, filename string,
	filesize int64, fileAddr string) bool {

	stmt, err := mydb.DBconn().Prepare(
		"insert ignore into tbl_file (`file_sha1`,`file_name`,`file_size`," +
			"`file_addr`,`status`) values (?,?,?,?,1)")
	if err != nil {
		fmt.Printf("failed to prepare statement , err : %s\n", err)
		return false
	}
	defer stmt.Close()

	res, err := stmt.Exec(fileHash, filename, filesize, fileAddr)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := res.RowsAffected(); err == nil {
		if rf <= 0 {
			fmt.Printf("File with hash : %s has been uploaded before", fileHash)
		}
		return true
	}
	return false
}

type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

//	GetFileMeta : 从mysql获取文件元信息;
func GetFileMeta(fileHash string) (*TableFile, error) {
	stmt, err := mydb.DBconn().Prepare(
		"select file_sha1,file_addr,file_name,file_size from tbl_file " +
			"where file_sha1=? and status=1 limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	tFile := TableFile{}
	err = stmt.QueryRow(fileHash).Scan(
		&tFile.FileHash, &tFile.FileAddr, &tFile.FileName, &tFile.FileSize)
	if err != nil {
		if err == sql.ErrNoRows {
			// 查不到对应记录， 返回参数及错误均为nil
			return nil, nil
		} else {
			fmt.Println(err.Error())
			return nil, err
		}
		//fmt.Println( err.Error() )
		//return nil , err
	}
	return &tFile, err
}

// GetFileMetaList : 从mysql批量获取文件元信息
func GetFileMetaList(limit int) ([]TableFile, error) {
	stmt, err := mydb.DBconn().Prepare(
		"select file_sha1,file_addr,file_name,file_size from tbl_file " +
			"where status=1 limit ?")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(limit)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	cloumns, _ := rows.Columns()
	values := make([]sql.RawBytes, len(cloumns))
	var tfiles []TableFile
	for i := 0; i < len(values) && rows.Next(); i++ {
		tfile := TableFile{}
		err = rows.Scan(&tfile.FileHash, &tfile.FileAddr,
			&tfile.FileName, &tfile.FileSize)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		tfiles = append(tfiles, tfile)
	}
	fmt.Println(len(tfiles))
	return tfiles, nil
}

// UpdateFileLocation : 更新文件的存储地址(如文件被转移了)
func UpdateFileLocation(filehash string, fileaddr string) bool {
	stmt, err := mydb.DBconn().Prepare(
		"update tbl_file set`file_addr`=? where  `file_sha1`=? limit 1")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(fileaddr, filehash)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			fmt.Printf("更新文件location失败, filehash:%s", filehash)
		}
		return true
	}
	return false
}
