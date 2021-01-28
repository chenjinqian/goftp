package main

import (
       "fmt"
       "io/ioutil"
       "github.com/aWildProgrammer/fconf"
       "github.com/dutchcoders/goftp"
       "os"
       "path/filepath"
       "path"
       // filepath.Join have bugs.
)

type ftpConfig struct {
     ip   string
     port string
     user string
     passwd string
     rpath string
     lpath string
     deepth int
     del    string
}

var fc = new(ftpConfig)
var c *fconf.Config
var ftp *goftp.FTP

func getSubDirs(rootPath string, num int)([]string, error){
     rlt_lst := []string{}
     err := getSubDirsAcc(&rlt_lst, rootPath, "", num)
     return rlt_lst, err
}

func getSubDirsAcc(rlt_lst *[]string, rootPath string,  dirPath string, num int)(error){
     // fmt.Println(dirPath)
     // TODO: not inplace parameter version function.
     if num == 0{
     	*rlt_lst = append(*rlt_lst, "")
	return nil
     }
     fileinfo, err := ioutil.ReadDir(path.Join(rootPath, dirPath))
     if err != nil{
     	// fmt.Println(dirPath)
     	panic(err)
     }
     for _, fi := range fileinfo{
	 if fi.IsDir(){
	    if num > 1{
	       if fi.IsDir(){
	       	  getSubDirsAcc(rlt_lst, rootPath, path.Join(dirPath, fi.Name()), num - 1)
	       	  // if (dirPath == ""){
		  //    getSubDirsAcc(rlt_lst, rootPath, fi.Name(), num - 1)
		  //    // TODO: filepath.join
		  //    }else{
		  //    getSubDirsAcc(rlt_lst, rootPath, dirPath + "/" + fi.Name(), num - 1)
		  //    }
		  }
	    }else{
		// if (dirPath == ""){
		//    tmp := fi.Name()
		//    *rlt_lst = append(*rlt_lst, tmp)
		// }else{
		// 	tmp := dirPath + "/" + fi.Name()
		// 	*rlt_lst = append(*rlt_lst, tmp)
		// }				
		tmp := path.Join(dirPath, fi.Name())
		*rlt_lst = append(*rlt_lst, tmp)
	    }
	 }
     }
     return nil
}

func getSubFiles(fp string)([]string){
     fileinfo, err := ioutil.ReadDir(fp)
     if err != nil{
     	panic(err)
     }
     rlt_lst := []string{} //  {} necceary?
     for _, fi := range fileinfo{
     	 if fi.IsDir(){	    
	 }else {
	       rlt_lst = append(rlt_lst, fi.Name())
	 }
     }
     return rlt_lst
}

func initConfig(){
     var err error
     c, err = fconf.NewFileConf("config.ini")
     if err != nil{
     	panic(err)
     }
     fc.ip = c.String("ftp.ip")
     fc.port = c.String("ftp.port")
     fc.user = c.String("ftp.user")
     fc.passwd = c.String("ftp.passwd")
     fc.rpath = c.String("ftp.rpath")
     fc.lpath = c.String("ftp.lpath")
     fc.deepth, _ = c.Int("ftp.deepth")
     fc.del= c.String("ftp.del")
     // TODO: debug, fconf can not read second part, if not ended with empty line.
}

// DONE, read config file
// ftpUploadDir
// ftpCwd
// ftpUpload
// shutil.delete

// func ftpUpload(ftpObj goftp , localFile []string, remoteFile []string)(error){
// var ftp *goftp.FTP
//   if ftp, err = goftp.Connect("serverip:port"); err != nil {
//    fmt.Println(err)
//    }
// }

func delFileEmptyDir(fp string)(error){
     // del file and empty folder
     err := os.Remove(fp)
     if err != nil{
     	fmt.Println("delete not success")
     }else{
	fmt.Println("deleted " + fp)
     }
     return err
}

func delDirRec(fp string)(error){
     fileinfo, err := ioutil.ReadDir(fp)
     if err != nil{
     	panic(err)
     }
     for _, fi := range fileinfo{
     	 if fi.IsDir(){
	    delDirRec(fp + "/" + fi.Name())
	 }else{
		delFileEmptyDir(fp + "/" + fi.Name())
	 }
     }
     delFileEmptyDir(fp)
     return nil
}

func ftpLogin(fc ftpConfig)(*goftp.FTP, error){
     ftp, err := goftp.Connect(fc.ip + ":" + fc.port)
     if err != nil{
     	fmt.Println("ftp connect fail")
	panic(err)
     }
     err = ftp.Login(fc.user, fc.passwd)
     if err != nil{
     	fmt.Println("ftp login fail")
	panic(err)
     }
     err = ftp.Cwd(fc.rpath)
     // err = ftp.Cwd("/log/ZHHN")
     if err != nil{
     	fmt.Println("login fail")
     }
     print("login success")
     return ftp, err
}



func ftpUploadDir(ftp *goftp.FTP, lfp string, rfp string, delFlag string)(error){
     // fmt.Println("#156ftp upload dir local")
     _, dname := filepath.Split(lfp)
     //fmt.Println(lfp, "#168remote", rfp, "name", dname, "flag", delFlag)
     // TODO: should check local dir exists.
     err := ftp.Cwd(rfp)
     if err != nil{
     	fmt.Println("#164, cwd failed: ", rfp)
	return err
     }
     err = ftp.Cwd(dname)
     // TODO: should use path.split and save one parameter slot.
     if err != nil{
         err = ftp.Mkd(dname)
	 if err != nil{
	    fmt.Println("#177 cwd fail:", dname)
	    fmt.Println("#172, create folder fail:", dname)
	    return err
	 }
	 err = ftp.Cwd(dname)
	 if err != nil{
	    fmt.Println("#189, after create folder, fail to cwd: ", dname)
	    return err
	 }
     }
     subFiles := getSubFiles(lfp)
     for _, subFile := range subFiles {
     	 err = ftp.Upload(path.Join(lfp, subFile))
	 if err != nil{
	    fmt.Println("#184, upload failed:", path.Join(lfp, subFile))
	 }else{
     	    fmt.Println("#165 uploaded", subFile)
	    if delFlag == "del"{
	       delFileEmptyDir(path.Join(lfp, subFile))
	    }
	 }
     }
     sub_dirs, err := getSubDirs(lfp, 1)
     for _, subDirOne := range sub_dirs{
     	 err = ftpUploadDir(ftp, path.Join(lfp, subDirOne), (rfp + "/" + dname),  delFlag)
	 if err != nil{
	    return err
	 }
     }
     if delFlag == "del"{
     	delFileEmptyDir(lfp)
     }
     return nil
}

func main_acc()(error){
     initConfig()
     rlt_lst, err := getSubDirs(fc.lpath, fc.deepth)
     if err != nil{
     	fmt.Println("")
     	fmt.Println(err)
     	return err
     }
     for _, rlt := range rlt_lst{
     	 fmt.Println(rlt)
     }
     // // delFileEmptyDir("./todel")
     // file_lst := getSubFiles(".")
     // for _, fi := range file_lst{
     // 	 fmt.Println(fi)
     // }
     // fmt.Println("deleting dir")
     // delDirRec("todel")
     // _, filename := filepath.Split("/tmp/test3.txt")
     // fmt.Println("filename: ", filename)
     // // upload and delete local files.
     ftp, err = ftpLogin(*fc)
     if err != nil{
     	panic(err)
     }
     for _, rltDirOne := range rlt_lst{
     	 err = ftpUploadDir(ftp, path.Join(fc.lpath, rltDirOne), fc.rpath, fc.del)
	 if err != nil{
	    return err
	 }
     }
     defer ftp.Close()
     return nil
}

func main() {
     // rlt_lst := []string{}
     // // read config into fc 
     // fmt.Println(fc.ip)
     // fmt.Println(fc.port)
     // fmt.Println(fc.user)
     // fmt.Println(fc.passwd)
     // fmt.Println(fc.lpath)
     // fmt.Println(">>> root dir <<<")
     err := main_acc()
     if err != nil{
     	fmt.Println("error happend, restart main", err)
     	main()
     }
     return
}
