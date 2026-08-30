package knowledge_test

import (
	"goraven/core/knowledge"
	unit_test "goraven/util/unit"
	"testing"
)

func TestConvertFile(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	err := knowledge.ConvertFile("123.pdf", "123.md")
	if err != nil {
		panic(err)
	}
}

func TestChunkFile(t *testing.T) {
	unitTest := unit_test.GetUnitTest()
	unitTest.Run()

	err := knowledge.ConvertFile("123.pdf", "123.txt")
	if err != nil {
		panic(err)
	}

}
