package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// SkillTool 实现动态技能加载工具
type SkillTool struct {
	baseDir string
}

// NewSkillTool 创建一个新的技能工具
func NewSkillTool(baseDir string) tool.InvokableTool {
	return &SkillTool{baseDir: baseDir}
}

type skillMatter struct {
	Name        string
	Description string
}

func (t *SkillTool) listSkills() ([]skillMatter, error) {
	entries, err := os.ReadDir(t.baseDir)
	if err != nil {
		return nil, err
	}

	var matters []skillMatter
	for _, entry := range entries {
		if entry.IsDir() {
			skillDir := filepath.Join(t.baseDir, entry.Name())
			skillFile := filepath.Join(skillDir, "SKILL.md")
			content, err := os.ReadFile(skillFile)
			if err != nil {
				continue
			}

			// 简单的元数据解析 (解析 YAML 格式的 description)
			matter := skillMatter{Name: entry.Name()}
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "description:") {
					matter.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
					// 移除可能存在的引号
					matter.Description = strings.Trim(matter.Description, "\"'")
					break
				}
			}
			matters = append(matters, matter)
		}
	}
	return matters, nil
}

// Info 返回工具信息，包含所有可用技能的清单
func (t *SkillTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	matters, err := t.listSkills()
	if err != nil {
		// 如果目录不存在或读取失败，返回一个空的基础信息
		return &schema.ToolInfo{
			Name: "skill",
			Desc: "加载技能工具（当前无可用技能）",
		}, nil
	}

	desc := "按需加载预定义技能锦囊。当用户的问题或你的任务匹配某个技能的描述时，调用此工具加载该技能的完整专业指令。\n\n可用技能清单："
	for _, m := range matters {
		desc += fmt.Sprintf("\n- %s: %s", m.Name, m.Description)
	}
	desc += "\n\n调用方式：传入 skill 参数为技能名称。加载后，你将获得该领域的专家指南。"

	return &schema.ToolInfo{
		Name: "skill",
		Desc: desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"skill": {
				Type:     schema.String,
				Desc:     "技能名称，必须是清单中的名称",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行技能加载逻辑
func (t *SkillTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Skill string `json:"skill"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	if args.Skill == "" {
		return "", fmt.Errorf("必须指定技能名称")
	}

	// 安全检查：防止路径穿越
	skillName := filepath.Base(args.Skill)
	skillFile := filepath.Join(t.baseDir, skillName, "SKILL.md")
	
	content, err := os.ReadFile(skillFile)
	if err != nil {
		return "", fmt.Errorf("找不到技能 [%s] 的指令集，请确认名称是否正确", args.Skill)
	}

	return string(content), nil
}
