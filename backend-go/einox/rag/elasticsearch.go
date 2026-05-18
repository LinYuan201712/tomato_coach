package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/tomato/backend/config"
)

type ElasticsearchStore struct {
	client *elasticsearch.Client
	index  string
}

func NewElasticsearchStore(cfg *config.ElasticConfig) (*ElasticsearchStore, error) {
	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}
	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, err
	}

	// 增加连接检查：尝试获取服务器信息
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("无法连接到 Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch 返回错误: %s", res.String())
	}

	return &ElasticsearchStore{
		client: client,
		index:  cfg.Index,
	}, nil
}

func (s *ElasticsearchStore) Add(ctx context.Context, docs []*Document) error {
	for _, doc := range docs {
		data, err := json.Marshal(doc)
		if err != nil {
			return err
		}

		req := esapi.IndexRequest{
			Index:      s.index,
			DocumentID: doc.ID,
			Body:       bytes.NewReader(data),
			Refresh:    "true",
		}

		res, err := req.Do(ctx, s.client)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.IsError() {
			return fmt.Errorf("failed to index document %s: %s", doc.ID, res.String())
		}
	}
	return nil
}

func (s *ElasticsearchStore) Search(ctx context.Context, vector []float64, topK int, threshold float64) ([]*Document, error) {
	// 默认向量搜索
	query := map[string]any{
		"knn": map[string]any{
			"field":          "vector",
			"query_vector":   vector,
			"k":              topK,
			"num_candidates": 100,
		},
	}

	return s.doSearch(ctx, query, topK)
}

func (s *ElasticsearchStore) HybridSearch(ctx context.Context, queryStr string, vector []float64, topK int, threshold float64, weight float64) ([]*Document, error) {
	// ES 8.x 支持混合搜索 (Hybrid Search)
	// 使用 RRF 或者直接在 query 中组合

	// 简单的加权组合示例 (需要 ES 支持或者手动 RRF)
	// 这里我们使用 ES 的 hybrid 搜索语法 (如果版本支持) 或者组合搜索

	body := map[string]any{
		"size": topK,
		"query": map[string]any{
			"bool": map[string]any{
				"should": []map[string]any{
					{
						"match": map[string]any{
							"content": map[string]any{
								"query": queryStr,
								"boost": 1.0 - weight, // BM25 权重
							},
						},
					},
				},
			},
		},
		"knn": map[string]any{
			"field":          "vector",
			"query_vector":   vector,
			"k":              topK,
			"num_candidates": 100,
			"boost":          weight, // 向量权重
		},
	}

	docs, err := s.doSearch(ctx, body, topK)
	if err == nil {
		return docs, nil
	}

	// Some local Elasticsearch builds do not support the top-level knn object in
	// _search. Keep knowledge retrieval usable by falling back to BM25.
	fmt.Printf("[RAG] HybridSearch failed, fallback to BM25: %v\n", err)
	fallback := map[string]any{
		"size": topK,
		"query": map[string]any{
			"match": map[string]any{
				"content": queryStr,
			},
		},
	}
	return s.doSearch(ctx, fallback, topK)
}

func (s *ElasticsearchStore) GetFullDocument(ctx context.Context, fileName string, userID int64) (string, error) {
	// 聚合该文件的所有分片并按顺序拼接
	// 注意：分片 ID 生成方式需要支持排序，或者元数据中有序号
	// 这里简单实现：搜索该文件名的所有分片
	query := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must": []map[string]any{
					{"term": map[string]any{"metadata.file_name.keyword": fileName}},
					{"term": map[string]any{"metadata.user_id": userID}},
				},
			},
		},
		"sort": []map[string]any{
			{"id.keyword": "asc"}, // 假设 ID 后缀是按位置生成的
		},
		"size": 1000,
	}

	docs, err := s.doSearch(ctx, query, 1000)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, doc := range docs {
		sb.WriteString(doc.Content)
	}
	return sb.String(), nil
}

func (s *ElasticsearchStore) doSearch(ctx context.Context, body map[string]any, topK int) ([]*Document, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(s.index),
		s.client.Search.WithBody(&buf),
		s.client.Search.WithTrackScores(true),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]any
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]any)["type"], e["error"].(map[string]any)["reason"])
	}

	var r map[string]any
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	hits := r["hits"].(map[string]any)["hits"].([]any)
	results := make([]*Document, 0, len(hits))
	for _, hit := range hits {
		source := hit.(map[string]any)["_source"]
		score := hit.(map[string]any)["_score"].(float64)

		data, _ := json.Marshal(source)
		var doc Document
		json.Unmarshal(data, &doc)
		doc.Score = score
		results = append(results, &doc)
	}

	return results, nil
}

func (s *ElasticsearchStore) List(ctx context.Context) ([]*Document, error) {
	return s.doSearch(ctx, map[string]any{"query": map[string]any{"match_all": map[string]any{}}}, 1000)
}

func (s *ElasticsearchStore) Delete(ctx context.Context, id string) error {
	res, err := s.client.Delete(s.index, id, s.client.Delete.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}
