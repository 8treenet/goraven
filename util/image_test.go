package util

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func makeTestJpeg(t *testing.T, w, h, quality int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			img.SetRGBA(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}))
	return buf.Bytes()
}

func TestCompressImageBytes(t *testing.T) {
	// 小图：不处理，原样返回
	small := makeTestJpeg(t, 800, 600, 90)
	out, changed, err := CompressImageBytes(small, ImageMaxEdge)
	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, small, out)

	// 大图：缩小重编码，最长边不超过 ImageMaxEdge
	large := makeTestJpeg(t, 4000, 3000, 100)
	out, changed, err = CompressImageBytes(large, ImageMaxEdge)
	require.NoError(t, err)
	require.True(t, changed)
	require.Less(t, len(out), len(large))

	img, _, err := image.Decode(bytes.NewReader(out))
	require.NoError(t, err)
	b := img.Bounds()
	require.Equal(t, 2048, b.Dx())
	require.LessOrEqual(t, b.Dy(), 2048)

	// 非图片字节：解码失败，返回错误与原字节
	bad := []byte("not an image")
	out, changed, err = CompressImageBytes(bad, ImageMaxEdge)
	require.Error(t, err)
	require.False(t, changed)
	require.Equal(t, bad, out)
}

func TestCompressImage(t *testing.T) {
	dir := t.TempDir()

	// 大图：生成缩略副本，返回副本路径
	src := filepath.Join(dir, "big.jpg")
	require.NoError(t, os.WriteFile(src, makeTestJpeg(t, 4000, 3000, 100), 0o644))
	dst := filepath.Join(dir, "big_thumb.jpg")
	used, err := CompressImage(src, dst, ImageMaxEdge)
	require.NoError(t, err)
	require.Equal(t, dst, used)
	_, err = os.Stat(dst)
	require.NoError(t, err)

	// 小图：不生成副本，返回原路径
	smallSrc := filepath.Join(dir, "small.jpg")
	require.NoError(t, os.WriteFile(smallSrc, makeTestJpeg(t, 800, 600, 90), 0o644))
	smallDst := filepath.Join(dir, "small_thumb.jpg")
	used, err = CompressImage(smallSrc, smallDst, ImageMaxEdge)
	require.NoError(t, err)
	require.Equal(t, smallSrc, used)
	_, err = os.Stat(smallDst)
	require.True(t, os.IsNotExist(err))

	// 非图片：解码失败，返回原路径与错误，不生成副本
	badSrc := filepath.Join(dir, "bad.jpg")
	require.NoError(t, os.WriteFile(badSrc, []byte("not an image"), 0o644))
	badDst := filepath.Join(dir, "bad_thumb.jpg")
	used, err = CompressImage(badSrc, badDst, ImageMaxEdge)
	require.Error(t, err)
	require.Equal(t, badSrc, used)
	_, err = os.Stat(badDst)
	require.True(t, os.IsNotExist(err))
}
