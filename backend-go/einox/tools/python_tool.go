package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// PythonParams Python 执行参数
type PythonParams struct {
	Code string `json:"code" desc:"要执行的 Python 代码"`
}

// NewPythonTool 创建一个真实的 Python 解释器工具
// 用于执行技能中定义的逻辑校准、数据分析或科学计算。
func NewPythonTool() tool.BaseTool {
	return utils.NewTool(
		&schema.ToolInfo{
			Name: "python_interpreter",
			Desc: "执行真实的 Python 代码。适用于情感熵分析、数据拟合、复杂逻辑验证等任务。返回代码的标准输出 (Stdout)。",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"code": {
					Type: "string",
					Desc: "要运行的完整 Python 代码字符串",
				},
			}),
		},
		func(ctx context.Context, params *PythonParams) (string, error) {
			if params.Code == "" {
				return "错误: 未提供任何代码", nil
			}

			// 1. 创建临时 Python 脚本文件
			tmpFile, err := os.CreateTemp("", "tomato_exec_*.py")
			if err != nil {
				return "", fmt.Errorf("无法创建临时文件: %v", err)
			}
			// 确保函数结束时删除临时文件
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(params.Code); err != nil {
				return "", fmt.Errorf("写入临时文件失败: %v", err)
			}
			tmpFile.Close()

			// 2. 准备执行命令
			// 优先尝试 python，如果失败可以根据环境配置改为 python3
			cmd := exec.CommandContext(ctx, "python", tmpFile.Name())
			
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			// 3. 运行并捕获结果
			err = cmd.Run()
			
			outStr := stdout.String()
			errStr := stderr.String()

			if err != nil {
				// 返回详细的错误信息，帮助 Agent 自我修正代码
				return fmt.Sprintf("Python 执行失败!\n错误信息: %v\n标准错误输出: %s\n标准输出: %s", err, errStr, outStr), nil
			}

			// 4. 返回标准输出结果
			if outStr == "" && errStr == "" {
				return "执行成功，但没有标准输出内容。", nil
			}
			
			return outStr, nil
		},
	)
}
