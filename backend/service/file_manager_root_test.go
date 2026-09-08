package service

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"

	"github.com/8treenet/freedom"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFileManagerRootOps(t *testing.T) {
	t.Run("mkdir and list use absolute root paths", func(t *testing.T) {
		root := t.TempDir()
		service := &FileManagerService{}

		if err := service.MkdirRoot(root, &vo.FileManagerMkdirReq{Path: "nested/dir"}); err != nil {
			t.Fatalf("MkdirRoot() error = %v", err)
		}
		if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("file"), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(root, ".hidden"), []byte("hidden"), 0644); err != nil {
			t.Fatalf("write hidden file: %v", err)
		}

		response, err := service.ListRoot(root, &vo.FileManagerListReq{})
		if err != nil {
			t.Fatalf("ListRoot() error = %v", err)
		}
		if len(response.Items) != 2 {
			t.Fatalf("ListRoot() returned %d items, want 2: %#v", len(response.Items), response.Items)
		}
		if response.Items[0].Name != "nested" || !response.Items[0].IsDir {
			t.Fatalf("ListRoot() first item = %#v, want nested directory", response.Items[0])
		}
		if response.Items[0].Path != filepath.Join(root, "nested") {
			t.Errorf("directory path = %q, want %q", response.Items[0].Path, filepath.Join(root, "nested"))
		}
		if response.Items[1].Name != "file.txt" || response.Items[1].Path != filepath.Join(root, "file.txt") {
			t.Errorf("file item = %#v", response.Items[1])
		}
	})

	t.Run("rejects traversal for root operations", func(t *testing.T) {
		root := t.TempDir()
		outside := filepath.Join(filepath.Dir(root), filepath.Base(root)+"-outside")
		if err := os.MkdirAll(outside, 0755); err != nil {
			t.Fatalf("create outside directory: %v", err)
		}
		t.Cleanup(func() { _ = os.RemoveAll(outside) })
		if err := os.WriteFile(filepath.Join(outside, "source.txt"), []byte("outside"), 0644); err != nil {
			t.Fatalf("write outside file: %v", err)
		}
		if err := writeRootTestZip(filepath.Join(root, "archive.zip"), "file.txt", []byte("archive")); err != nil {
			t.Fatalf("write archive: %v", err)
		}

		service := &FileManagerService{}
		if err := service.MkdirRoot(root, &vo.FileManagerMkdirReq{Path: "../created"}); err == nil {
			t.Fatal("MkdirRoot() accepted traversal")
		}
		if err := service.DeleteRoot(root, &vo.FileManagerDeleteReq{Paths: []string{"../source.txt"}}); err == nil {
			t.Fatal("DeleteRoot() accepted traversal")
		}
		if err := service.RenameRoot(root, &vo.FileManagerRenameReq{OldPath: "archive.zip", NewPath: "../renamed.zip"}); err == nil {
			t.Fatal("RenameRoot() accepted traversal")
		}
		if _, err := service.CompressRoot(root, &vo.FileManagerCompressReq{Paths: []string{"../source.txt"}, OutputName: "archive"}); err == nil {
			t.Fatal("CompressRoot() accepted traversal")
		}
		if err := service.DecompressRoot(root, &vo.FileManagerDecompressReq{Path: "../archive.zip"}); err == nil {
			t.Fatal("DecompressRoot() accepted traversal")
		}
		if _, _, err := service.DownloadRoot(root, "../source.txt"); err == nil {
			t.Fatal("DownloadRoot() accepted traversal")
		}
		if _, err := os.Stat(filepath.Join(outside, "source.txt")); err != nil {
			t.Fatalf("outside source was changed: %v", err)
		}
	})

	t.Run("accepts root-relative paths with a leading slash", func(t *testing.T) {
		root := t.TempDir()
		outside := t.TempDir()
		if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0644); err != nil {
			t.Fatalf("write outside file: %v", err)
		}

		service := &FileManagerService{}
		// 前端以 "/" 表示根目录，路径形如 "/dir"（根相对写法），必须等价于 "dir"。
		if err := service.MkdirRoot(root, &vo.FileManagerMkdirReq{Path: "/created"}); err != nil {
			t.Fatalf("MkdirRoot(/created) error = %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, "created")); err != nil {
			t.Fatalf("leading-slash directory not created: %v", err)
		}
		if _, err := service.ListRoot(root, &vo.FileManagerListReq{Dir: "/created"}); err != nil {
			t.Fatalf("ListRoot(/created) error = %v", err)
		}
		if err := service.MkdirRoot(root, &vo.FileManagerMkdirReq{Path: "/nested/dir"}); err != nil {
			t.Fatalf("MkdirRoot(/nested/dir) error = %v", err)
		}
		if err := service.DeleteRoot(root, &vo.FileManagerDeleteReq{Paths: []string{"/nested/dir"}}); err != nil {
			t.Fatalf("DeleteRoot(/nested/dir) error = %v", err)
		}

		// 绝对路径同样被归一化为根相对路径：操作落在根目录内，不会写穿到根目录之外。
		absoluteOutside := filepath.Join(outside, "absolute.txt")
		if err := service.MkdirRoot(root, &vo.FileManagerMkdirReq{Path: absoluteOutside}); err != nil {
			t.Fatalf("MkdirRoot(absolute) error = %v", err)
		}
		if _, err := os.Stat(filepath.Join(outside, "absolute.txt")); !os.IsNotExist(err) {
			t.Fatalf("absolute path escaped the root: %v", err)
		}
	})

	t.Run("rejects root-equivalent destructive paths", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "keep.txt"), []byte("keep"), 0644); err != nil {
			t.Fatalf("write keep file: %v", err)
		}
		if err := os.Mkdir(filepath.Join(root, "nested"), 0755); err != nil {
			t.Fatalf("create nested directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(root, "nested", "inner.txt"), []byte("inner"), 0644); err != nil {
			t.Fatalf("write inner file: %v", err)
		}

		service := &FileManagerService{}
		for _, path := range []string{"", ".", "foo/..", "./nested/.."} {
			if err := service.DeleteRoot(root, &vo.FileManagerDeleteReq{Paths: []string{path}}); err == nil {
				t.Fatalf("DeleteRoot() accepted root-equivalent path %q", path)
			}
		}
		if _, err := os.Stat(filepath.Join(root, "keep.txt")); err != nil {
			t.Fatalf("root contents did not survive rejected deletion: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, "nested", "inner.txt")); err != nil {
			t.Fatalf("root contents did not survive rejected deletion: %v", err)
		}

		if err := service.RenameRoot(root, &vo.FileManagerRenameReq{OldPath: "foo/..", NewPath: "bar/.."}); err == nil {
			t.Fatal("RenameRoot() accepted root-equivalent paths")
		}
		if err := service.RenameRoot(root, &vo.FileManagerRenameReq{OldPath: ".", NewPath: "renamed"}); err == nil {
			t.Fatal("RenameRoot() accepted a root-equivalent old path")
		}
		if err := service.RenameRoot(root, &vo.FileManagerRenameReq{OldPath: "keep.txt", NewPath: "."}); err == nil {
			t.Fatal("RenameRoot() accepted a root-equivalent new path")
		}
		if err := service.RenameRoot(root, &vo.FileManagerRenameReq{OldPath: "keep.txt", NewPath: "nested/.."}); err == nil {
			t.Fatal("RenameRoot() accepted a root-equivalent new path")
		}
		if _, err := os.Stat(filepath.Join(root, "keep.txt")); err != nil {
			t.Fatalf("root contents were changed by rejected renames: %v", err)
		}
	})

	t.Run("compresses at the root and preserves zip suffix", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, "source"), 0755); err != nil {
			t.Fatalf("create source directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(root, "source", "file.txt"), []byte("content"), 0644); err != nil {
			t.Fatalf("write source file: %v", err)
		}

		service := &FileManagerService{}
		response, err := service.CompressRoot(root, &vo.FileManagerCompressReq{
			Paths:      []string{"source"},
			OutputName: "bundle",
		})
		if err != nil {
			t.Fatalf("CompressRoot() error = %v", err)
		}
		if response.ZipPath != "bundle.zip" {
			t.Fatalf("ZipPath = %q, want bundle.zip", response.ZipPath)
		}
		if _, err := os.Stat(filepath.Join(root, "bundle.zip")); err != nil {
			t.Fatalf("root-level archive missing: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, "source", "bundle.zip")); !os.IsNotExist(err) {
			t.Fatalf("archive was not kept at the root, stat error = %v", err)
		}

		if err := service.DeleteRoot(root, &vo.FileManagerDeleteReq{Paths: []string{"source"}}); err != nil {
			t.Fatalf("DeleteRoot() error = %v", err)
		}
		if err := service.DecompressRoot(root, &vo.FileManagerDecompressReq{Path: "bundle.zip"}); err != nil {
			t.Fatalf("DecompressRoot() error = %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, "source", "file.txt")); err != nil {
			t.Fatalf("decompressed source missing: %v", err)
		}
	})

	t.Run("compresses leading-slash root-relative paths", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("<html></html>"), 0644); err != nil {
			t.Fatalf("write source file: %v", err)
		}
		if err := os.Mkdir(filepath.Join(root, "docs"), 0755); err != nil {
			t.Fatalf("create docs directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(root, "docs", "readme.md"), []byte("readme"), 0644); err != nil {
			t.Fatalf("write nested file: %v", err)
		}

		service := &FileManagerService{}
		response, err := service.CompressRoot(root, &vo.FileManagerCompressReq{
			Paths:      []string{"/index.html"},
			OutputName: "page",
		})
		if err != nil {
			t.Fatalf("CompressRoot() error = %v", err)
		}
		if response.ZipPath != "page.zip" {
			t.Fatalf("ZipPath = %q, want page.zip", response.ZipPath)
		}
		if _, err := os.Stat(filepath.Join(root, "page.zip")); err != nil {
			t.Fatalf("root-level archive missing: %v", err)
		}

		response, err = service.CompressRoot(root, &vo.FileManagerCompressReq{
			Paths:      []string{"/docs/readme.md"},
			OutputName: "docs",
		})
		if err != nil {
			t.Fatalf("CompressRoot(nested) error = %v", err)
		}
		if response.ZipPath != "docs.zip" {
			t.Fatalf("ZipPath = %q, want docs.zip", response.ZipPath)
		}
		if _, err := os.Stat(filepath.Join(root, "docs.zip")); err != nil {
			t.Fatalf("nested-source root archive missing: %v", err)
		}
	})

	t.Run("reports usage and downloads regular files", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, "dir"), 0755); err != nil {
			t.Fatalf("create directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(root, "one.txt"), []byte("one"), 0644); err != nil {
			t.Fatalf("write first file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(root, "dir", "two.txt"), []byte("two!"), 0644); err != nil {
			t.Fatalf("write second file: %v", err)
		}

		service := &FileManagerService{}
		usage, err := service.UsageRoot(root)
		if err != nil {
			t.Fatalf("UsageRoot() error = %v", err)
		}
		if usage.UsedSize != 7 || usage.FileCount != 2 {
			t.Fatalf("UsageRoot() = %#v, want 7 bytes and 2 files", usage)
		}

		path, name, err := service.DownloadRoot(root, filepath.Join("dir", "two.txt"))
		if err != nil {
			t.Fatalf("DownloadRoot() error = %v", err)
		}
		if path != filepath.Join(root, "dir", "two.txt") || name != "two.txt" {
			t.Fatalf("DownloadRoot() = %q, %q", path, name)
		}
		if _, _, err := service.DownloadRoot(root, "dir"); err == nil {
			t.Fatal("DownloadRoot() accepted a directory")
		}
	})

	t.Run("rejects symlink escapes", func(t *testing.T) {
		root := t.TempDir()
		outside := t.TempDir()
		if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0644); err != nil {
			t.Fatalf("write outside file: %v", err)
		}
		if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}

		service := &FileManagerService{}
		if err := service.MkdirRoot(root, &vo.FileManagerMkdirReq{Path: filepath.Join("link", "created")}); err == nil {
			t.Fatal("MkdirRoot() accepted a symlink escape")
		}
		if err := service.DeleteRoot(root, &vo.FileManagerDeleteReq{Paths: []string{filepath.Join("link", "secret.txt")}}); err == nil {
			t.Fatal("DeleteRoot() accepted a symlink escape")
		}
		if _, _, err := service.DownloadRoot(root, filepath.Join("link", "secret.txt")); err == nil {
			t.Fatal("DownloadRoot() accepted a symlink escape")
		}
	})
}

func TestFileManagerRootOpsUpload(t *testing.T) {
	service, db := newRootUploadService(t)

	t.Run("rejects wrong user", func(t *testing.T) {
		tempDir := t.TempDir()
		upload := createRootUpload(t, db, &po.ChunkUpload{
			UploadId: "root-upload-wrong-user",
			UserId:   "owner",
			FileName: "file.txt",
			TempDir:  tempDir,
			Status:   po.UploadStatusCompleted,
		})

		if _, err := service.UploadRoot(t.TempDir(), "other-user", &vo.FileManagerUploadReq{UploadId: upload.UploadId}); err == nil {
			t.Fatal("UploadRoot() accepted a different user")
		}
	})

	t.Run("rejects incomplete upload", func(t *testing.T) {
		tempDir := t.TempDir()
		upload := createRootUpload(t, db, &po.ChunkUpload{
			UploadId: "root-upload-incomplete",
			UserId:   "owner",
			FileName: "file.txt",
			TempDir:  tempDir,
			Status:   po.UploadStatusPending,
		})

		if _, err := service.UploadRoot(t.TempDir(), "owner", &vo.FileManagerUploadReq{UploadId: upload.UploadId}); err == nil {
			t.Fatal("UploadRoot() accepted an incomplete upload")
		}
	})

	t.Run("rejects unsafe filename and destination traversal", func(t *testing.T) {
		t.Run("filename", func(t *testing.T) {
			tempDir := t.TempDir()
			upload := createRootUpload(t, db, &po.ChunkUpload{
				UploadId: "root-upload-unsafe-name",
				UserId:   "owner",
				FileName: "../escape.txt",
				TempDir:  tempDir,
				Status:   po.UploadStatusCompleted,
			})

			if _, err := service.UploadRoot(t.TempDir(), "owner", &vo.FileManagerUploadReq{UploadId: upload.UploadId}); err == nil {
				t.Fatal("UploadRoot() accepted an unsafe filename")
			}
		})

		t.Run("destination", func(t *testing.T) {
			root := t.TempDir()
			tempDir := t.TempDir()
			upload := createRootUpload(t, db, &po.ChunkUpload{
				UploadId: "root-upload-unsafe-destination",
				UserId:   "owner",
				FileName: "file.txt",
				TempDir:  tempDir,
				Status:   po.UploadStatusCompleted,
			})
			if err := os.WriteFile(filepath.Join(tempDir, upload.FileName), []byte("content"), 0644); err != nil {
				t.Fatalf("write merged file: %v", err)
			}

			if _, err := service.UploadRoot(root, "owner", &vo.FileManagerUploadReq{
				UploadId: upload.UploadId,
				Dir:      "../outside",
			}); err == nil {
				t.Fatal("UploadRoot() accepted a traversal destination")
			}
			if _, err := os.Stat(filepath.Join(tempDir, upload.FileName)); err != nil {
				t.Fatalf("merged file was moved after rejected destination: %v", err)
			}
		})
	})

	t.Run("moves the merged file and consumes the upload", func(t *testing.T) {
		root := t.TempDir()
		tempDir := t.TempDir()
		upload := createRootUpload(t, db, &po.ChunkUpload{
			UploadId: "root-upload-success",
			UserId:   "owner",
			FileName: "file.txt",
			TempDir:  tempDir,
			Status:   po.UploadStatusCompleted,
		})
		contents := []byte("merged content")
		if err := os.WriteFile(filepath.Join(tempDir, upload.FileName), contents, 0644); err != nil {
			t.Fatalf("write merged file: %v", err)
		}

		response, err := service.UploadRoot(root, "owner", &vo.FileManagerUploadReq{
			UploadId: upload.UploadId,
			Dir:      "/nested/files", // 前端根相对写法（"/" 开头）必须与 "nested/files" 等价
		})
		if err != nil {
			t.Fatalf("UploadRoot() error = %v", err)
		}
		if response.Path != filepath.Join("/nested/files", upload.FileName) {
			t.Fatalf("response path = %q, want %q", response.Path, filepath.Join("/nested/files", upload.FileName))
		}

		got, err := os.ReadFile(filepath.Join(root, response.Path))
		if err != nil {
			t.Fatalf("read moved file: %v", err)
		}
		if string(got) != string(contents) {
			t.Fatalf("moved file = %q, want %q", got, contents)
		}
		if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
			t.Fatalf("temp directory still exists, stat error = %v", err)
		}

		var stored po.ChunkUpload
		if err := db.First(&stored, "upload_id = ?", upload.UploadId).Error; err != nil {
			t.Fatalf("reload upload: %v", err)
		}
		if stored.Status != po.UploadStatusUsed {
			t.Fatalf("upload status = %d, want used (%d)", stored.Status, po.UploadStatusUsed)
		}
	})

	t.Run("rejects symlinked destination", func(t *testing.T) {
		root := t.TempDir()
		tempDir := t.TempDir()
		upload := createRootUpload(t, db, &po.ChunkUpload{
			UploadId: "root-upload-symlink-destination",
			UserId:   "owner",
			FileName: "file.txt",
			TempDir:  tempDir,
			Status:   po.UploadStatusCompleted,
		})
		if err := os.WriteFile(filepath.Join(tempDir, upload.FileName), []byte("content"), 0644); err != nil {
			t.Fatalf("write merged file: %v", err)
		}
		outside := t.TempDir()
		outsideFile := filepath.Join(outside, "target.txt")
		if err := os.WriteFile(outsideFile, []byte("original"), 0644); err != nil {
			t.Fatalf("write outside file: %v", err)
		}
		if err := os.Symlink(outsideFile, filepath.Join(root, upload.FileName)); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}

		if _, err := service.UploadRoot(root, "owner", &vo.FileManagerUploadReq{UploadId: upload.UploadId}); err == nil {
			t.Fatal("UploadRoot() wrote through a symlink")
		}
		got, err := os.ReadFile(outsideFile)
		if err != nil {
			t.Fatalf("read outside file: %v", err)
		}
		if string(got) != "original" {
			t.Fatalf("outside file = %q, want %q", got, "original")
		}
		if _, err := os.Stat(filepath.Join(tempDir, upload.FileName)); err != nil {
			t.Fatalf("merged file was consumed after rejected upload: %v", err)
		}
	})

	t.Run("fails when temp directory cleanup fails", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("running as root: permission-based cleanup failure is not reproducible")
		}
		root := t.TempDir()
		tempDir := t.TempDir()
		upload := createRootUpload(t, db, &po.ChunkUpload{
			UploadId: "root-upload-cleanup-failure",
			UserId:   "owner",
			FileName: "file.txt",
			TempDir:  tempDir,
			Status:   po.UploadStatusCompleted,
		})
		if err := os.WriteFile(filepath.Join(tempDir, upload.FileName), []byte("content"), 0644); err != nil {
			t.Fatalf("write merged file: %v", err)
		}
		locked := filepath.Join(tempDir, "locked")
		if err := os.Mkdir(locked, 0755); err != nil {
			t.Fatalf("create locked directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(locked, "held.txt"), []byte("held"), 0644); err != nil {
			t.Fatalf("write held file: %v", err)
		}
		if err := os.Chmod(locked, 0500); err != nil {
			t.Fatalf("lock directory: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(locked, 0700) })

		if _, err := service.UploadRoot(root, "owner", &vo.FileManagerUploadReq{UploadId: upload.UploadId}); err == nil {
			t.Fatal("UploadRoot() ignored temp directory cleanup failure")
		}
		if _, err := os.Stat(filepath.Join(root, upload.FileName)); err != nil {
			t.Fatalf("uploaded file missing after failed upload: %v", err)
		}
	})

	t.Run("fails when marking the upload used fails", func(t *testing.T) {
		root := t.TempDir()
		tempDir := t.TempDir()
		upload := createRootUpload(t, db, &po.ChunkUpload{
			UploadId: "root-upload-mark-failure",
			UserId:   "owner",
			FileName: "file.txt",
			TempDir:  tempDir,
			Status:   po.UploadStatusCompleted,
		})
		if err := os.WriteFile(filepath.Join(tempDir, upload.FileName), []byte("content"), 0644); err != nil {
			t.Fatalf("write merged file: %v", err)
		}
		if err := db.Exec(`CREATE TRIGGER fail_mark_upload_used BEFORE UPDATE ON chunk_upload BEGIN SELECT RAISE(ABORT, 'forced update failure'); END;`).Error; err != nil {
			t.Fatalf("create trigger: %v", err)
		}
		t.Cleanup(func() {
			if err := db.Exec(`DROP TRIGGER IF EXISTS fail_mark_upload_used`).Error; err != nil {
				t.Errorf("drop trigger: %v", err)
			}
		})

		if _, err := service.UploadRoot(root, "owner", &vo.FileManagerUploadReq{UploadId: upload.UploadId}); err == nil {
			t.Fatal("UploadRoot() ignored MarkUploadUsed failure")
		}
		if _, err := os.Stat(filepath.Join(root, upload.FileName)); err != nil {
			t.Fatalf("uploaded file missing after failed upload: %v", err)
		}
		var stored po.ChunkUpload
		if err := db.First(&stored, "upload_id = ?", upload.UploadId).Error; err != nil {
			t.Fatalf("reload upload: %v", err)
		}
		if stored.Status != po.UploadStatusCompleted {
			t.Fatalf("upload status = %d, want unchanged completed (%d)", stored.Status, po.UploadStatusCompleted)
		}
	})
}

func newRootUploadService(t *testing.T) (*FileManagerService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite memory db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sqlite db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&po.ChunkUpload{}); err != nil {
		t.Fatalf("migrate chunk uploads: %v", err)
	}

	unitTest := freedom.NewUnitTest()
	unitTest.InstallDB(func() interface{} { return db })
	unitTest.Run()

	var repo *repository.HFSRepository
	unitTest.FetchRepository(&repo)
	return &FileManagerService{HFSRepo: repo}, db
}

func createRootUpload(t *testing.T, db *gorm.DB, upload *po.ChunkUpload) *po.ChunkUpload {
	t.Helper()
	if err := db.Create(upload).Error; err != nil {
		t.Fatalf("create upload: %v", err)
	}
	return upload
}

func writeRootTestZip(path, name string, contents []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create(name)
	if err != nil {
		_ = file.Close()
		return err
	}
	if _, err := entry.Write(contents); err != nil {
		_ = writer.Close()
		_ = file.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
