package rag

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
)

var docxTextTagRE = regexp.MustCompile(`<w:t(?:\s[^>]*)?>([^<]*)</w:t>`)

// IsDocx 判断是否为 Word 2007+ 文档
func IsDocx(fileName string) bool {
	return strings.HasSuffix(strings.ToLower(fileName), ".docx")
}

// IsDoc 判断是否为旧版 Word 二进制文档
func IsDoc(fileName string) bool {
	return strings.HasSuffix(strings.ToLower(fileName), ".doc")
}

// IsMarkdown 判断是否为 Markdown
func IsMarkdown(fileName string) bool {
	ext := strings.ToLower(fileName)
	return strings.HasSuffix(ext, ".md") || strings.HasSuffix(ext, ".markdown")
}

// IsBinaryGarbledText 判断内容是否像误当作 UTF-8 的二进制（docx/zip 等）
func IsBinaryGarbledText(s string) bool {
	if s == "" {
		return false
	}
	if strings.HasPrefix(s, "PK\x03\x04") {
		return true
	}
	if strings.HasPrefix(s, "\xD0\xCF\x11\xE0") {
		return true
	}

	sample := s
	if len(sample) > 4096 {
		sample = sample[:4096]
	}
	bad := 0
	runes := 0
	for _, r := range sample {
		runes++
		if r == '\n' || r == '\r' || r == '\t' {
			continue
		}
		if r < 32 && r != '\t' {
			bad++
			continue
		}
		if r == utf8.RuneError {
			bad++
		}
	}
	if runes == 0 {
		return true
	}
	return float64(bad)/float64(runes) > 0.12
}

// ExtractText 从上传文件字节中提取可索引/预览的纯文本
func ExtractText(fileName string, content []byte) (string, error) {
	if len(content) == 0 {
		return "", fmt.Errorf("文件内容为空")
	}

	switch {
	case IsDocx(fileName):
		return extractDocxText(content)
	case IsDoc(fileName):
		return "", fmt.Errorf("不支持旧版 .doc 格式，请在 Word 中另存为 .docx 后重新上传")
	case IsPDF(fileName):
		return "", fmt.Errorf("PDF 应走 MinerU 解析流程")
	default:
		return decodeTextContent(content), nil
	}
}

func decodeTextContent(content []byte) string {
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		content = content[3:]
	}
	if len(content) >= 2 && content[0] == 0xFF && content[1] == 0xFE {
		return decodeUTF16LE(content[2:])
	}
	if utf8.Valid(content) {
		return string(content)
	}
	return string(bytes.ToValidUTF8(content, []byte("?")))
}

func extractDocxText(content []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("无效的 docx 文件（非 ZIP 格式）: %w", err)
	}

	var docXML []byte
	for _, f := range zr.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		docXML, err = io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return "", err
		}
		break
	}
	if len(docXML) == 0 {
		return "", fmt.Errorf("docx 中未找到 word/document.xml")
	}

	docXML = normalizeXMLEncoding(docXML)

	text, err := parseWordDocumentXML(docXML)
	if err != nil {
		return "", err
	}
	text = strings.TrimSpace(text)
	if text == "" || IsBinaryGarbledText(text) {
		text = extractDocxTextByRegex(docXML)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("docx 中未提取到文本内容")
	}
	if IsBinaryGarbledText(text) {
		return "", fmt.Errorf("docx 解析结果异常，请确认文件未损坏")
	}
	return text, nil
}

func normalizeXMLEncoding(data []byte) []byte {
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		return []byte(decodeUTF16LE(data[2:]))
	}
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		return []byte(decodeUTF16BE(data[2:]))
	}
	return data
}

func decodeUTF16BE(b []byte) string {
	if len(b)%2 != 0 {
		b = b[:len(b)-1]
	}
	runes := make([]rune, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		runes = append(runes, rune(b[i])<<8|rune(b[i+1]))
	}
	return string(runes)
}

func decodeUTF16LE(b []byte) string {
	if len(b)%2 != 0 {
		b = b[:len(b)-1]
	}
	runes := make([]rune, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		runes = append(runes, rune(b[i])|rune(b[i+1])<<8)
	}
	return string(runes)
}

func extractDocxTextByRegex(xml []byte) string {
	matches := docxTextTagRE.FindAllStringSubmatch(string(xml), -1)
	if len(matches) == 0 {
		return ""
	}
	var buf strings.Builder
	lastPara := false
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		part := strings.TrimSpace(m[1])
		if part == "" {
			continue
		}
		if buf.Len() > 0 && !lastPara {
			buf.WriteByte(' ')
		}
		buf.WriteString(part)
		lastPara = false
	}
	return collapseBlankLines(buf.String())
}

func parseWordDocumentXML(data []byte) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var buf strings.Builder
	inText := false
	inParagraph := false

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("解析 docx XML 失败: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "t", "delText", "instrText":
				inText = true
			case "p":
				if inParagraph && buf.Len() > 0 {
					buf.WriteByte('\n')
				}
				inParagraph = true
			case "tab":
				buf.WriteByte('\t')
			case "br":
				buf.WriteByte('\n')
			}
		case xml.EndElement:
			if t.Name.Local == "p" {
				if inParagraph && buf.Len() > 0 {
					buf.WriteByte('\n')
				}
				inParagraph = false
			}
			if t.Name.Local == "t" || t.Name.Local == "delText" || t.Name.Local == "instrText" {
				inText = false
			}
		case xml.CharData:
			if inText {
				buf.Write(t)
			}
		}
	}

	return collapseBlankLines(buf.String()), nil
}

func collapseBlankLines(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" && len(out) > 0 && out[len(out)-1] == "" {
			continue
		}
		out = append(out, line)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

// PickPreviewText 从分片中还原可预览正文，跳过旧的二进制乱码分片
func PickPreviewText(docs []*Document) string {
	for _, d := range docs {
		if d == nil || d.MetaData == nil {
			continue
		}
		if ft, ok := d.MetaData[MetaFullText].(string); ok {
			ft = strings.TrimSpace(ft)
			if ft != "" && !IsBinaryGarbledText(ft) {
				return ft
			}
		}
	}

	var sb strings.Builder
	for _, d := range docs {
		if d == nil || IsBinaryGarbledText(d.Content) {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(strings.TrimSpace(d.Content))
	}
	return strings.TrimSpace(sb.String())
}
