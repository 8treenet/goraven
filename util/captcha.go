package util

import (
	"crypto/rand"
	"math/big"
	"strconv"

	"github.com/mojocn/base64Captcha"
)

var captchaDigitDriver = base64Captcha.NewDriverDigit(60, 100, 2, 0.6, 60)

func RandCaptchaPair() (int, int, error) {
	ai, err := rand.Int(rand.Reader, big.NewInt(28))
	if err != nil {
		return 0, 0, err
	}
	a := int(ai.Int64()) + 1

	bMaxN := 30 - a
	if bMaxN > 29 {
		bMaxN = 29
	}
	bi, err := rand.Int(rand.Reader, big.NewInt(int64(bMaxN)))
	if err != nil {
		return 0, 0, err
	}
	b := int(bi.Int64()) + 1
	return a, b, nil
}

func RenderDigitBase64(num int) (string, error) {
	item, err := captchaDigitDriver.DrawCaptcha(strconv.Itoa(num))
	if err != nil {
		return "", err
	}
	return item.EncodeB64string(), nil
}
