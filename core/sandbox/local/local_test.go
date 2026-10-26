package local_test

import (
	"goraven/core/sandbox"
	"testing"
)

func TestLocalSandbox_Upload(t *testing.T) {
	box, boxerr := sandbox.NewSandbox("999")
	if boxerr != nil {
		panic(boxerr)
	}
	t.Log(box.Upload("/Users/ys/go/src/github.com/8treenet/goraven/core/sandbox", "/Users/ys/work/config/goraven/user/999/temp/sandbox"))
	t.Log(box.Upload("/Users/ys/go/src/github.com/8treenet/goraven/core/sandbox/sandbox.go", "/Users/ys/work/config/goraven/user/999/temp/xxx123.go"))
}

func TestLocalSandbox_Download(t *testing.T) {
	box, boxerr := sandbox.NewSandbox("999")
	if boxerr != nil {
		panic(boxerr)
	}
	t.Log(box.Download("/Users/ys/work/config/goraven/user/999/temp/xxx123.go"))
}

func TestFileManager_List(t *testing.T) {
	box, boxerr := sandbox.NewSandbox("999")
	if boxerr != nil {
		panic(boxerr)
	}
	t.Log(box.NewFileManager().List("/", "name", "desc"))
}

func TestFileManager_Mkdir(t *testing.T) {
	box, boxerr := sandbox.NewSandbox("999")
	if boxerr != nil {
		panic(boxerr)
	}
	t.Log(box.NewFileManager().Mkdir("test123"))
}

func TestFileManager_Rename(t *testing.T) {
	box, boxerr := sandbox.NewSandbox("999")
	if boxerr != nil {
		panic(boxerr)
	}
	t.Log(box.NewFileManager().Rename("/test123", "/test234"))
}

func TestFileManager_Compress(t *testing.T) {
	box, boxerr := sandbox.NewSandbox("999")
	if boxerr != nil {
		panic(boxerr)
	}
	t.Log(box.NewFileManager().Compress([]string{"/documents/go_concurrency_plan_2026-04-25_47291.md", "/documents/go_concurrency_plan_20260425_82731.md"}, "test.zip"))
}

func TestFileManager_Decompress(t *testing.T) {
	box, boxerr := sandbox.NewSandbox("999")
	if boxerr != nil {
		panic(boxerr)
	}
	t.Log(box.NewFileManager().Decompress("/test.zip", false))
}

func TestFileManager_Delete(t *testing.T) {
	box, boxerr := sandbox.NewSandbox("999")
	if boxerr != nil {
		panic(boxerr)
	}
	t.Log(box.NewFileManager().Delete([]string{"/test.zip"}))
}

func TestFileManager_GetUsage(t *testing.T) {
	box, boxerr := sandbox.NewSandbox("999")
	if boxerr != nil {
		panic(boxerr)
	}
	t.Log(box.NewFileManager().GetUsage())
}

func TestLocalSandbox_GetStorageUsage(t *testing.T) {
	box, boxerr := sandbox.NewSandbox("90a431bee756432492c134f510bad949")
	if boxerr != nil {
		panic(boxerr)
	}
	t.Log(box.NewFileManager().GetUsage())
}

func TestLocalSandbox_View(t *testing.T) {
	box, _ := sandbox.NewSandbox("admin123")
	t.Log(box.GetWorkspace())
}
