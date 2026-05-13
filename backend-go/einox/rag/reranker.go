package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
)

// AliyunReranker 阿里云 DashScope Rerank 实现
type AliyunReranker struct {
	APIKey string
	Model  string
}

func NewAliyunReranker(apiKey string) *AliyunReranker {
	return &AliyunReranker{
		APIKey: apiKey,
		Model:  "gte-rerank-v2",
	}
}

func (r *AliyunReranker) Rerank(ctx context.Context, query string, docs []*Document) ([]*Document, error) {
	if len(docs) == 0 {
		return docs, nil
	}

	// 准备输入
	docTexts := make([]string, len(docs))
	for i, doc := range docs {
		docTexts[i] = doc.Content
	}

	payload := map[string]any{
		"model": r.Model,
		"input": map[string]any{
			"query":     query,
			"documents": docTexts,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://dashscope.aliyuncs.com/api/v1/services/rerank/text-rerank/text-rerank", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+r.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("aliyun rerank failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Output struct {
			Results []struct {
				Index          int     `json:"index"`
				RelevanceScore float64 `json:"relevance_score"`
			} `json:"results"`
		} `json:"output"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 将分数写回原文档
	for _, res := range result.Output.Results {
		if res.Index < len(docs) {
			docs[res.Index].Score = res.RelevanceScore
		}
	}

	// 按重排序后的分数排序
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Score > docs[j].Score
	})

	return docs, nil
}
