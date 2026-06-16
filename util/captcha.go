package util

import (
	"crypto/rand"
	"math/big"
	"strconv"

	"github.com/mojocn/base64Captcha"
)

// captchaDigitDriver 复用同一份字体配置生成单个加数图片（最多 2 位数字）
var captchaDigitDriver = base64Captcha.NewDriverDigit(60, 100, 2, 0.6, 60)

// RandCaptchaPair 生成两个 1..29 之间且和 ≤30 的加数，供登录算术验证码使用
func RandCaptchaPair() (int, int, error) {
	ai, err := rand.Int(rand.Reader, big.NewInt(28))
	if err != nil {
		return 0, 0, err
	}
	a := int(ai.Int64()) + 1 // 1..28

	bMaxN := 30 - a
	if bMaxN > 29 {
		bMaxN = 29
	}
	bi, err := rand.Int(rand.Reader, big.NewInt(int64(bMaxN)))
	if err != nil {
		return 0, 0, err
	}
	b := int(bi.Int64()) + 1 // 1..bMaxN
	return a, b, nil
}

// RenderDigitBase64 把单个数字渲染成图片，返回带 data:image/png;base64, 前缀的字符串
// 通过 DrawCaptcha(content) 强制图片内容为 num，避免 driver 内部随机
func RenderDigitBase64(num int) (string, error) {
	item, err := captchaDigitDriver.DrawCaptcha(strconv.Itoa(num))
	if err != nil {
		return "", err
	}
	return item.EncodeB64string(), nil
}
