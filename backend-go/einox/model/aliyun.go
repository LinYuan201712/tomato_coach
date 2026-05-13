package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tomato/backend/config"
)

// AliyunEmbedder 阿里云 DashScope 文本嵌入模型
type AliyunEmbedder struct {
	apiKey string
	model  string
	client *http.Client
}

func NewAliyunEmbedder(ctx context.Context, cfg *config.AliyunConfig) (*AliyunEmbedder, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("阿里云 API 密钥为空")
	}
	return &AliyunEmbedder{
		apiKey: cfg.APIKey,
		model:  cfg.EmbeddingModel,
		client: &http.Client{},
	}, nil
}

// EmbedStrings 将一组文本转换为向量 (DashScope 接口，支持分批请求以绕过单次 10 条限制)
func (e *AliyunEmbedder) EmbedStrings(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	const batchSize = 10
	allEmbeddings := make([][]float64, len(texts))

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batchTexts := texts[i:end]
		batchRes, err := e.doEmbedBatch(ctx, batchTexts)
		if err != nil {
			return nil, err
		}

		// 将当前批次结果填入总结果集
		for j, emb := range batchRes {
			allEmbeddings[i+j] = emb
		}
	}

	return allEmbeddings, nil
}

type embedRequest struct {
	Model string `json:"model"`
	Input struct {
		Texts []string `json:"texts"`
	} `json:"input"`
	Parameters struct {
		TextType string `json:"text_type,omitempty"`
	} `json:"parameters,omitempty"`
}

type embedResponse struct {
	Output struct {
		Embeddings []struct {
			Embedding []float64 `json:"embedding"`
			TextIndex int       `json:"text_index"`
		} `json:"embeddings"`
	} `json:"output"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AliyunEmbedder) doEmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	reqBody := embedRequest{
		Model: e.model,
	}
	reqBody.Input.Texts = texts

	body, _ := json.Marshal(reqBody)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://dashscope.aliyuncs.com/api/v1/services/embeddings/text-embedding/text-embedding", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("阿里云 API 错误 [%d]: %s", resp.StatusCode, string(b))
	}

	var res embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	if res.Code != "" && res.Code != "0" {
		return nil, fmt.Errorf("阿里云 API 业务错误 [%s]: %s", res.Code, res.Message)
	}

	embs := make([][]float64, len(res.Output.Embeddings))
	for _, emb := range res.Output.Embeddings {
		if emb.TextIndex < len(embs) {
			embs[emb.TextIndex] = emb.Embedding
		}
	}
	return embs, nil
}
