package util

import (
	"bytes"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"os"

	_ "golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// ImageMaxEdge 图片缩放目标最长边像素。
// 多模态模型按像素面积计 token，缩小尺寸可显著减少 token 消耗与传输体积。
const ImageMaxEdge = 2048

// imageJpegQuality 重编码 JPEG 质量（1-100）
const imageJpegQuality = 85

// CompressImageBytes 对图片字节做缩小压缩，重新编码为 JPEG 返回。
// 最长边不超过 maxEdge 时不处理，返回原始字节与 false；解码失败返回原始字节与错误。
// 注意：重编码会丢失 EXIF（含方向信息），竖拍照片可能旋转，需调用方权衡。
func CompressImageBytes(data []byte, maxEdge int) ([]byte, bool, error) {
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data, false, err
	}
	bounds := src.Bounds()
	if bounds.Dx() <= maxEdge && bounds.Dy() <= maxEdge {
		return data, false, nil
	}

	dst := scaleImage(src, maxEdge)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: imageJpegQuality}); err != nil {
		return data, false, err
	}
	return buf.Bytes(), true, nil
}

// CompressImage 将 srcPath 图片缩小压缩为 JPEG 副本写入 dstPath（目标目录需已存在）。
// 返回调用方应使用的路径：成功缩放时返回 dstPath；源图无需缩放、解码或写入失败时
// 不生成副本，返回 srcPath。
func CompressImage(srcPath, dstPath string, maxEdge int) (string, error) {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return srcPath, err
	}
	out, changed, err := CompressImageBytes(data, maxEdge)
	if err != nil || !changed {
		return srcPath, err
	}
	if err := os.WriteFile(dstPath, out, 0o644); err != nil {
		return srcPath, err
	}
	return dstPath, nil
}

// scaleImage 将图片按最长边缩放到 maxEdge（保持宽高比）
func scaleImage(src image.Image, maxEdge int) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	longest := w
	if h > longest {
		longest = h
	}
	scale := float64(maxEdge) / float64(longest)
	nw := max(1, int(float64(w)*scale))
	nh := max(1, int(float64(h)*scale))

	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)
	return dst
}
