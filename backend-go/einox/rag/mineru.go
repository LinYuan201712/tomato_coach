package rag

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tomato/backend/config"
	mu "github.com/opendatalab/MinerU-Ecosystem/sdk/go"
)

// MinerUWorker 封装 MinerU SDK 调用
type MinerUWorker struct {
	cfg *config.MinerUConfig
}

func NewMinerUWorker(cfg *config.MinerUConfig) *MinerUWorker {
	return &MinerUWorker{cfg: cfg}
}

// Extract 将 PDF 提取为 Markdown 字符串
func (w *MinerUWorker) Extract(ctx context.Context, source string) (string, error) {
	if w.cfg == nil || !w.cfg.Enabled {
		return "", fmt.Errorf("MinerU is disabled")
	}
	if w.cfg.Token == "" {
		return "", fmt.Errorf("MinerU token is missing")
	}

	client, err := mu.New(w.cfg.Token)
	if err != nil {
		return "", fmt.Errorf("failed to create MinerU client: %w", err)
	}

	poll, _ := time.ParseDuration(w.cfg.PollTimeout)
	if poll == 0 {
		poll = 15 * time.Minute
	}

	opts := []mu.ExtractOption{
		mu.WithOCR(true),
		mu.WithPollTimeout(poll),
	}

	// 这里的 source 可以是本地文件路径或 URL
	result, err := client.Extract(ctx, source, opts...)
	if err != nil {
		return "", fmt.Errorf("MinerU extract error: %w", err)
	}
	if err := result.Err(); err != nil {
		return "", fmt.Errorf("MinerU task failed: %w", err)
	}

	return result.Markdown, nil
}

// IsPDF 判断文件名是否为 PDF
func IsPDF(fileName string) bool {
	return strings.HasSuffix(strings.ToLower(fileName), ".pdf")
}
