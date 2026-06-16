package test

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"raven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

var (
	token = "rvn_2d2291c9fe6b49998e8d9b310f076e31"
)

func TestHfsPublicDownload(t *testing.T) {
	linkId := "test-link-id"
	respBody, resp := requests.NewHTTPRequest(domain + "/api/hfs/public/" + linkId).Get().ToBytes()
	if resp.Error != nil {
		t.Log("error:", resp.Error)
		return
	}
	t.Log("public download response:", respBody)
}

func TestHfsPrivateDownload(t *testing.T) {
	respBody, resp := requests.NewHTTPRequest(domain+"/api/hfs/private?path=/documents/go_concurrency_plan_2026-04-25_47291.md").
		Get().
		SetHeaderValue("Authorization", "Bearer "+"rvn_8931423dce414667bff8ddfcb3b11ff3").ToBytes()
	if resp.Error != nil {
		t.Log("error:", resp.Error)
		return
	}
	fmt.Println(resp)
	os.WriteFile("/Users/ys/text.md", respBody, 0644)
}

func TestHfsCreateUpload(t *testing.T) {
	filepath := "/Users/ys/temp/doctest/3.pptx"

	if filepath == "" {
		t.Skip("请设置 filepath 变量")
	}

	fileInfo, err := os.Stat(filepath)
	if err != nil {
		t.Fatalf("获取文件信息失败: %v", err)
	}
	fileSize := fileInfo.Size()
	fileName := fileInfo.Name()

	chunkSize := 512 * 1024
	totalChunks := int(fileSize) / chunkSize
	if int(fileSize)%chunkSize != 0 {
		totalChunks++
	}

	req := vo.ChunkUploadCreateReq{
		FileName:	fileName,
		FileSize:	fileSize,
		ChunkSize:	chunkSize,
		TotalChunks:	totalChunks,
	}

	bodystr, httpresp := requests.NewHTTPRequest(domain+"/api/hfs/upload/create").Post().SetJSONBody(req).SetHeaderValue("Authorization", "Bearer "+token).ToString()
	if httpresp.Error != nil {
		t.Fatalf("创建上传任务失败: %v", httpresp.Error)
	}
	t.Log(bodystr)
}

func TestHfsUploadChunk(t *testing.T) {
	filepath := "/Users/ys/temp/doctest/3.pptx"
	uploadId := "3ecc09a8f5164af99afcb0c8ab230311"
	chunkSize := 512 * 1024

	if filepath == "" || uploadId == "" {
		t.Skip("请设置 filepath 和 uploadId 变量")
	}

	fileInfo, err := os.Stat(filepath)
	if err != nil {
		t.Fatalf("获取文件信息失败: %v", err)
	}
	fileName := fileInfo.Name()
	totalChunks := int(fileInfo.Size()) / chunkSize
	if int(fileInfo.Size())%chunkSize != 0 {
		totalChunks++
	}

	file, err := os.Open(filepath)
	if err != nil {
		t.Fatalf("打开文件失败: %v", err)
	}
	defer file.Close()

	client := &http.Client{}
	for i := 0; i < totalChunks; i++ {
		offset := int64(i) * int64(chunkSize)
		file.Seek(offset, 0)

		buffer := make([]byte, chunkSize)
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			t.Fatalf("读取文件失败: %v", err)
		}

		chunkData := buffer[:n]
		hash := md5.Sum(chunkData)
		chunkMd5 := hex.EncodeToString(hash[:])

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", fileName)
		if err != nil {
			t.Fatalf("创建表单文件失败: %v", err)
		}
		part.Write(chunkData)
		writer.Close()

		url := fmt.Sprintf("%s/api/hfs/upload/chunk?uploadId=%s&chunkIndex=%d&chunkMd5=%s",
			domain, uploadId, i, chunkMd5)

		req, err := http.NewRequest("PUT", url, &body)
		if err != nil {
			t.Fatalf("创建请求失败: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("上传分片 %d 失败: %v", i, err)
		}
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var chunkRsp struct {
			Code	int	`json:"code"`
			Msg	string	`json:"msg"`
		}
		json.Unmarshal(respBody, &chunkRsp)
		if chunkRsp.Code != 0 {
			t.Fatalf("上传分片 %d 失败: code=%d, msg=%s", i, chunkRsp.Code, chunkRsp.Msg)
		}
		t.Logf("上传分片 %d/%d 成功, md5=%s", i+1, totalChunks, chunkMd5)
	}
	t.Logf("所有分片上传完成, 共 %d 个分片", totalChunks)
}

func TestHfsMergeUpload(t *testing.T) {
	uploadId := "3ecc09a8f5164af99afcb0c8ab230311"

	if uploadId == "" {
		t.Skip("请设置 uploadId 变量")
	}

	req := vo.ChunkMergeReq{
		UploadId: uploadId,
	}

	var rsp struct {
		Code	int			`json:"code"`
		Msg	string			`json:"msg"`
		Data	vo.ChunkMergeRsp	`json:"data,omitempty"`
	}

	resp := requests.NewHTTPRequest(domain+"/api/hfs/upload/merge").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if resp.Error != nil {
		t.Fatalf("合并分片失败: %v", resp.Error)
	}
	t.Logf("合并分片成功: filePath=%s, fileName=%s, fileSize=%d", rsp.Data.FilePath, rsp.Data.FileName, rsp.Data.FileSize)
}

func TestHfsAssets(t *testing.T) {
	uploadId := "87919ed89fed4df2a7315474303b3020"

	if uploadId == "" {
		t.Skip("请设置 uploadId 变量")
	}

	req := vo.AssetsReq{
		UploadId: uploadId,
	}
	var rsp struct {
		Code	int		`json:"code"`
		Msg	string		`json:"msg"`
		Data	vo.AssetsRsp	`json:"data,omitempty"`
	}

	resp := requests.NewHTTPRequest(domain+"/api/hfs/assets").
		Post().
		SetJSONBody(req).
		SetHeaderValue("Authorization", "Bearer "+token).
		ToJSON(&rsp)
	if resp.Error != nil {
		t.Fatalf("提交静态资源失败: %v", resp.Error)
	}
	t.Logf("提交静态资源成功: path=%s", rsp.Data.Path)
}
