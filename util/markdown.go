package util

import (
	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

func ConvertMarkDown(html string) (string, error) {
	return htmltomarkdown.ConvertString(html)
}
