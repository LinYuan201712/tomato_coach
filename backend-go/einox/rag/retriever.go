package rag

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/cloudwego/eino/schema"
)

type Retriever struct {
	store     VectorStore
	embedder  Embedder
	topK      int
	threshold float64
}

type RetrieverConfig struct {
	Store     VectorStore
	Embedder  Embedder
	TopK      int
	Threshold float64
}

func NewRetriever(cfg *RetrieverConfig) *Retriever {
	topK := cfg.TopK
	if topK <= 0 {
		topK = 5
	}
	threshold := cfg.Threshold
	if threshold <= 0 {
		threshold = 0.5
	}
	return &Retriever{
		store:     cfg.Store,
		embedder:  cfg.Embedder,
		topK:      topK,
		threshold: threshold,
	}
}

func (r *Retriever) Retrieve(ctx context.Context, query string) ([]*Document, error) {
	vectors, err := r.embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("嵌入查询失败: %w", err)
	}
	if len(vectors) == 0 || len(vectors[0]) == 0 {
		return nil, fmt.Errorf("向量为空")
	}
	docs, err := r.store.Search(ctx, vectors[0], r.topK, r.threshold)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}
	return docs, nil
}

type RAGPipeline struct {
	retriever *Retriever
	splitter  DocumentSplitter
}

func NewRAGPipeline(retriever *Retriever, splitter DocumentSplitter) *RAGPipeline {
	return &RAGPipeline{retriever, splitter}
}

func (p *RAGPipeline) IndexDocuments(ctx context.Context, docs []*Document) error {
	if p.splitter != nil {
		var err error
		docs, err = p.splitter.Split(ctx, docs)
		if err != nil {
			return fmt.Errorf("切分失败: %w", err)
		}
	}
	texts := make([]string, len(docs))
	for i, doc := range docs {
		texts[i] = doc.Content
	}
	vectors, err := p.retriever.embedder.EmbedStrings(ctx, texts)
	if err != nil {
		return fmt.Errorf("嵌入失败: %w", err)
	}
	for i, doc := range docs {
		doc.Vector = vectors[i]
	}
	return p.retriever.store.Add(ctx, docs)
}

func (p *RAGPipeline) Query(ctx context.Context, query string) (*RAGResult, error) {
	docs, err := p.retriever.Retrieve(ctx, query)
	if err != nil {
		return nil, err
	}
	context := BuildContext(docs)
	return &RAGResult{
		Query:        query,
		Documents:    docs,
		Context:      context,
		HasKnowledge: len(docs) > 0,
	}, nil
}

type RAGResult struct {
	Query        string
	Documents    []*Document
	Context      string
	HasKnowledge bool
}

func BuildContext(docs []*Document) string {
	if len(docs) == 0 {
		return "(未找到相关知识)"
	}

	// 按文件分组
	fileMap := make(map[string][]*Document)
	fileOrder := make([]string, 0)

	for _, doc := range docs {
		fileName, _ := doc.MetaData[MetaFileName].(string)
		if fileName == "" {
			fileName = "未知文件"
		}
		if _, ok := fileMap[fileName]; !ok {
			fileOrder = append(fileOrder, fileName)
		}
		fileMap[fileName] = append(fileMap[fileName], doc)
	}

	var builder strings.Builder
	builder.WriteString("以下是检索到的参考资料。请在回答时引用它们，并在文末列出参考资料清单，格式为 [n] 来源：[[文件名]]。如果你没有使用这些资料，请不要胡编乱造来源。\n\n")
	totalChars := 0
	const maxContextChars = 6000 // 调高单次知识库字符上限

	for i, fileName := range fileOrder {
		if totalChars >= maxContextChars {
			break
		}

		fileDocs := fileMap[fileName]
		// 1. 按起始索引排序
		sort.Slice(fileDocs, func(i, j int) bool {
			si, _ := fileDocs[i].MetaData[MetaStartIndex].(int)
			sj, _ := fileDocs[j].MetaData[MetaStartIndex].(int)
			return si < sj
		})

		// 2. 合并区间
		mergedChunks := make([]string, 0)
		if len(fileDocs) > 0 {
			currentContent := fileDocs[0].Content
			_, _ = fileDocs[0].MetaData[MetaStartIndex].(int)
			currentEnd, _ := fileDocs[0].MetaData[MetaEndIndex].(int)

			for j := 1; j < len(fileDocs); j++ {
				nextStart, _ := fileDocs[j].MetaData[MetaStartIndex].(int)
				nextEnd, _ := fileDocs[j].MetaData[MetaEndIndex].(int)
				nextContent := fileDocs[j].Content

				// 如果有重叠或相邻 (允许 1 个字符的缝隙作为相邻处理)
				if nextStart <= currentEnd {
					// 如果 nextEnd 超过 currentEnd，则追加不重叠的部分
					if nextEnd > currentEnd {
						overlap := currentEnd - nextStart
						if overlap < 0 {
							overlap = 0
						}
						// 注意：这里的索引是 rune 索引
						nextRunes := []rune(nextContent)
						if overlap < len(nextRunes) {
							currentContent += string(nextRunes[overlap:])
						}
						currentEnd = nextEnd
					}
					// 如果 nextEnd <= currentEnd，说明完全包含在内，跳过
				} else {
					// 不连续，保存当前段落，开启新段落
					mergedChunks = append(mergedChunks, currentContent)
					currentContent = nextContent
					currentEnd = nextEnd
				}
			}
			mergedChunks = append(mergedChunks, currentContent)
		}

		// 3. 写入 Builder
		builder.WriteString(fmt.Sprintf("[Ref %d] (来源: %s)\n", i+1, fileName))
		for _, content := range mergedChunks {
			if totalChars+len(content) > maxContextChars {
				remaining := maxContextChars - totalChars
				if remaining > 0 {
					// 粗略截断
					builder.WriteString(content[:remaining])
					builder.WriteString("... [由于长度限制被省略]")
				}
				totalChars = maxContextChars
				break
			}
			builder.WriteString(content)
			builder.WriteString("\n")
			totalChars += len(content)
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

type SimpleRAG struct {
	store        VectorStore
	embedder     Embedder
	splitter     DocumentSplitter
	chatModel    RewriterModel // 用于查询重写
	topK         int
	hybridWeight float64
	reranker     Reranker
}

type RewriterModel interface {
	Generate(ctx context.Context, msgs []*schema.Message) (*schema.Message, error)
}

type Reranker interface {
	Rerank(ctx context.Context, query string, docs []*Document) ([]*Document, error)
}

func NewSimpleRAG(store VectorStore, embedder Embedder, chatModel RewriterModel, topK int, hybridWeight float64) *SimpleRAG {
	return &SimpleRAG{
		store:        store,
		embedder:     embedder,
		splitter:     NewRecursiveSplitter(300, 50), // 调整为约 300 字符（对应约 150-200 token）
		chatModel:    chatModel,
		topK:         topK,
		hybridWeight: hybridWeight,
	}
}

func (r *SimpleRAG) SetReranker(reranker Reranker) {
	r.reranker = reranker
}

type LLMReranker struct {
	model RewriterModel
}

func NewLLMReranker(model RewriterModel) *LLMReranker {
	return &LLMReranker{model: model}
}

func (r *LLMReranker) Rerank(ctx context.Context, query string, docs []*Document) ([]*Document, error) {
	if len(docs) <= 1 {
		return docs, nil
	}

	// 简单的 LLM 打分逻辑示例
	// 实际项目中可能需要更复杂的 Prompt
	return docs, nil // 暂时返回原结果，留作扩展接口
}

func (r *SimpleRAG) AddText(ctx context.Context, text string, metadata map[string]any) error {
	doc := &Document{
		ID:       GenerateID(text),
		Content:  text,
		MetaData: metadata,
	}
	docs, err := r.splitter.Split(ctx, []*Document{doc})
	if err != nil {
		return err
	}
	texts := make([]string, len(docs))
	for i, d := range docs {
		texts[i] = d.Content
	}
	vectors, err := r.embedder.EmbedStrings(ctx, texts)
	if err != nil {
		return err
	}
	if len(vectors) != len(docs) {
		return fmt.Errorf("向量不匹配: 期望 %d, 实际 %d", len(docs), len(vectors))
	}
	for i, d := range docs {
		d.Vector = vectors[i]
	}
	err = r.store.Add(ctx, docs)
	if err != nil {
		return err
	}

	// 异步保存数据
	if ms, ok := r.store.(*MemoryVectorStore); ok {
		go ms.Save("data/rag_store.json")
	}
	return nil
}

func (r *SimpleRAG) Query(ctx context.Context, query string) (string, []*Document, error) {
	vectors, err := r.embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return "", nil, err
	}

	var docs []*Document
	if hs, ok := r.store.(HybridStore); ok {
		// 使用混合搜索
		docs, err = hs.HybridSearch(ctx, query, vectors[0], r.topK, 0.3, r.hybridWeight)
	} else {
		// 降级为纯向量搜索
		docs, err = r.store.Search(ctx, vectors[0], r.topK, 0.3)
	}

	if err != nil {
		return "", nil, err
	}
	return BuildContext(docs), docs, nil
}

func (r *SimpleRAG) ProfessionalQuery(ctx context.Context, query string, timeNow string, knowledgeBase string, userID int64) (string, []*Document, error) {
	if r.chatModel == nil {
		return "", nil, fmt.Errorf("RAG model not initialized")
	}

	// 1. 提取核心要点
	extractMsgs := []*schema.Message{
		{Role: schema.System, Content: "你是一个文本分析专家。请从用户的查询中提取核心检索词，直接返回检索词，不要有任何额外说明。"},
		{Role: schema.User, Content: query},
	}
	extractResp, err := r.chatModel.Generate(ctx, extractMsgs)
	coreQuery := query
	if err == nil {
		coreQuery = strings.TrimSpace(extractResp.Content)
	}

	usedKeywords := make([]string, 0)
	allDocs := make([]*Document, 0)
	docMap := make(map[string]bool)

	// 执行最多 2 轮查询重写与检索
	for i := 0; i < 2; i++ {
		currentQuery := coreQuery
		if i > 0 {
			// 第二轮才进行复杂的重写
			rewriteMsgs, err := r.prepareRewriteMessages(ctx, query, timeNow, knowledgeBase, usedKeywords)
			if err == nil {
				resp, err := r.chatModel.Generate(ctx, rewriteMsgs)
				if err == nil {
					currentQuery = strings.TrimSpace(resp.Content)
				}
			}
		}

		usedKeywords = append(usedKeywords, currentQuery)
		fmt.Printf("[DEBUG] Round %d Search Query: %s\n", i+1, currentQuery)

		// 2. 检索
		_, docs, err := r.Query(ctx, currentQuery)
		if err == nil {
			for _, doc := range docs {
				if !docMap[doc.ID] {
					allDocs = append(allDocs, doc)
					docMap[doc.ID] = true
				}
			}
		}

		if len(allDocs) >= r.topK*2 {
			break
		}
	}

	// 4. 重排序
	if r.reranker != nil && len(allDocs) > 0 {
		reranked, err := r.reranker.Rerank(ctx, query, allDocs)
		if err == nil {
			allDocs = reranked
		}
	}

	// 5. 构建上下文（移除激进的全文替换逻辑，改回纯切片模式）
	finalContext := BuildContext(allDocs)
	return finalContext, allDocs, nil
}

func (r *SimpleRAG) prepareRewriteMessages(ctx context.Context, query string, timeNow string, knowledgeBase string, used []string) ([]*schema.Message, error) {
	fromTemplate := struct {
		QueryRewriting interface {
			Format(ctx context.Context, vars map[string]any) ([]*schema.Message, error)
		}
	}{
		// 这是一个 hack，因为 prompt 包和 rag 包循环引用
		// 在实际代码中，我们会通过某种方式注入模板
	}
	_ = fromTemplate

	// 由于包依赖问题，我们在这里直接构建消息，或者假设 prompt 模板已经通过某种方式暴露
	// 这里我们使用一个临时的简单实现，实际项目中应通过接口或 DI 解决
	usedStr := strings.Join(used, ", ")
	if usedStr == "" {
		usedStr = "无"
	}

	systemPrompt := fmt.Sprintf(`你非常擅长于使用rag进行数据检索，你的目标是在充分理解用户的问题后进行向量化检索。
现在时间：%s
你要优化并提取搜索的查询内容。请遵循以下规则重写查询内容：
- 根据用户的问题和上下文，重写应该进行搜索的关键词
- 如果需要使用时间，则根据当前时间给出需要查询的具体时间日期信息
- 保持查询简洁，查询内容通常不超过3个关键词, 最多不要超过5个关键词
- 参考Elasticsearch搜索查询习惯重写关键字。
- 直接返回优化后的搜索词，不要有任何额外说明。
- 尽量不要使用下面这些已使用过的关键词，因为之前使用这些关键词搜索到的结果不符合预期，已使用过的关键词：%s
- 尽量不使用知识库名字《%s》中包含的关键词`, timeNow, usedStr, knowledgeBase)

	return []*schema.Message{
		{Role: schema.System, Content: systemPrompt},
		{Role: schema.User, Content: query},
	}, nil
}

func (r *SimpleRAG) ListDocuments(ctx context.Context, userID int64) ([]*Document, error) {
	docs, err := r.store.List(ctx)
	if err != nil {
		return nil, err
	}

	// 按源文件去重显示，并进行用户 ID 过滤
	seenSources := make(map[string]bool)
	userDocs := make([]*Document, 0)
	for _, doc := range docs {
		uid, ok := doc.MetaData["user_id"].(int64)
		if !ok {
			// 兼容不同类型的用户ID（有些地方可能是 float64，如果是 JSON 反序列化出来的）
			if fuid, ok := doc.MetaData["user_id"].(float64); ok {
				uid = int64(fuid)
			}
		}

		if uid == userID {
			sourceName, hasSource := doc.MetaData["file_name"].(string)
			if !hasSource {
				sourceName = "Unknown"
			}

			// 如果没见过这个源文件，或者是未知来源（每个单独显示），则添加
			if !seenSources[sourceName] || sourceName == "Unknown" {
				userDocs = append(userDocs, doc)
				seenSources[sourceName] = true
			}
		}
	}
	return userDocs, nil
}

func (r *SimpleRAG) DeleteDocument(ctx context.Context, docID string) error {
	return r.store.Delete(ctx, docID)
}

func (r *SimpleRAG) GetFullDocument(ctx context.Context, fileName string, userID int64) (string, error) {
	if hs, ok := r.store.(HybridStore); ok {
		return hs.GetFullDocument(ctx, fileName, userID)
	}

	// 如果不是 HybridStore，则通过 List 过滤（虽然效率稍低但能保证兼容性）
	docs, err := r.store.List(ctx)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	for _, d := range docs {
		// 鲁棒的文件名匹配
		metaFile, _ := d.MetaData[MetaFileName].(string)
		if strings.TrimSpace(metaFile) != strings.TrimSpace(fileName) {
			continue
		}

		// 鲁棒的用户 ID 匹配 (处理 float64 和 int64 混合情况)
		var uid int64
		if val, ok := d.MetaData["user_id"].(int64); ok {
			uid = val
		} else if val, ok := d.MetaData["user_id"].(float64); ok {
			uid = int64(val)
		}

		if uid == userID {
			sb.WriteString(d.Content)
			sb.WriteString("\n")
		}
	}
	return sb.String(), nil
}
