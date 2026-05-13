package rag

import (
	"strings"
	"testing"
)

func TestBuildContextMerging(t *testing.T) {
	docs := []*Document{
		{
			Content: "Hello world, this is a test.",
			MetaData: map[string]any{
				MetaFileName:   "test.txt",
				MetaStartIndex: 0,
				MetaEndIndex:   28,
			},
		},
		{
			Content: "is a test. Let's see if it merges.",
			MetaData: map[string]any{
				MetaFileName:   "test.txt",
				MetaStartIndex: 18,
				MetaEndIndex:   52,
			},
		},
		{
			Content: "see if it merges. It should be continuous.",
			MetaData: map[string]any{
				MetaFileName:   "test.txt",
				MetaStartIndex: 34,
				MetaEndIndex:   76,
			},
		},
		{
			Content: "Completely separate text.",
			MetaData: map[string]any{
				MetaFileName:   "test.txt",
				MetaStartIndex: 100,
				MetaEndIndex:   125,
			},
		},
	}

	context := BuildContext(docs)
	t.Logf("Generated Context:\n%s", context)

	// 预期合并后的文本包含：
	// "Hello world, this is a test. Let's see if it merges.It should be continuous."
	expectedPart1 := "Hello world, this is a test. Let's see if it merges.It should be continuous."
	if !strings.Contains(context, expectedPart1) {
		t.Errorf("Expected merged text not found.\nGot: %s", context)
	}

	if !strings.Contains(context, "Completely separate text.") {
		t.Errorf("Separate text not found.")
	}

	// 验证没有重复拼接导致的冗余（比如 "is a test. is a test."）
	if strings.Count(context, "is a test.") > 1 {
		t.Errorf("Redundant text found in context.")
	}
}
