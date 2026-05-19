package rag

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractText_Markdown(t *testing.T) {
	text, err := ExtractText("notes.md", []byte("# Hello\n\n世界"))
	require.NoError(t, err)
	assert.Contains(t, text, "Hello")
	assert.Contains(t, text, "世界")
}

func TestExtractText_Docx(t *testing.T) {
	docXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>Hello </w:t><w:t>World</w:t></w:r></w:p>
    <w:p><w:r><w:t>第二段</w:t></w:r></w:p>
  </w:body>
</w:document>`

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("word/document.xml")
	require.NoError(t, err)
	_, err = w.Write([]byte(docXML))
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	text, err := ExtractText("test.docx", buf.Bytes())
	require.NoError(t, err)
	assert.Contains(t, text, "Hello World")
	assert.Contains(t, text, "第二段")
}

func TestExtractText_DocRejected(t *testing.T) {
	_, err := ExtractText("legacy.doc", []byte{0xD0, 0xCF, 0x11, 0xE0})
	require.Error(t, err)
	assert.Contains(t, err.Error(), ".docx")
}

func TestIsBinaryGarbledText(t *testing.T) {
	assert.True(t, IsBinaryGarbledText("PK\x03\x04\x14\x00"))
	assert.False(t, IsBinaryGarbledText("程序分析实验报告模板"))
}

func TestPickPreviewText_SkipsBinaryChunks(t *testing.T) {
	docs := []*Document{
		{Content: "PK\x03\x04garbage", MetaData: map[string]any{MetaStartIndex: 0}},
		{Content: "正常正文", MetaData: map[string]any{MetaStartIndex: 100, MetaFullText: "正常正文第一段"}},
	}
	text := PickPreviewText(docs)
	assert.Equal(t, "正常正文第一段", text)
}
