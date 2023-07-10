package service

import (
	"fmt"
	"io"
	"main/utils"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/gin-gonic/gin"
)

func Upload(c *gin.Context) {
	w := c.Writer
	req := c.Request
	srcFile, head, err := req.FormFile("file")
	if err != nil {
		utils.RespFail(w, err.Error())
	}

	//获取文件名后缀
	// suffix := ".png"
	ofilName := head.Filename
	// fmt.Println("ofilName>>>>>", ofilName)
	// tem := strings.Split(ofilName, ".")
	// if len(tem) > 1 {
	// 	suffix = "." + tem[len(tem)-1]
	// }
	fileName := ""
	//前端录音文件固定传文件名为 blob
	if ofilName == "blob" {
		fileName = fmt.Sprintf("%d%04d%s", time.Now().Unix(), rand.Int31(), ".mp3")
	} else {
		ext := path.Ext(ofilName)
		fileName = fmt.Sprintf("%d%04d%s", time.Now().Unix(), rand.Int31(), ext)
	}

	//fileName := fmt.Sprintf("%d%04d%s", time.Now().Unix(), rand.Int31(), suffix)

	//创建文件
	dstFile, err := os.Create("./asset/upload/" + fileName)
	if err != nil {
		utils.RespFail(w, err.Error())
	}

	//复制文件到目标文件
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		utils.RespFail(w, err.Error())
	}

	//返回完整文件路径
	url := "./asset/upload/" + fileName
	utils.RespOK(w, url, "发送图片成功！！！")

}
