package tools

import (
	"github.com/cloudwego/eino/components/tool"
	"github.com/tomato/backend/einox/rag"
	"github.com/tomato/backend/internal/repository"
)

// GetAllTools 返回助手的全部可用工具
func GetAllTools(taskSvc TaskProvider, userRepo UserProvider, r *rag.SimpleRAG, reportRepo repository.StudyReportRepository) []tool.BaseTool {
	return []tool.BaseTool{
		NewTaskTool(taskSvc),
		NewPomodoroTool(),
		NewUserProfilingTool(userRepo),
		NewUpdateProfileTool(userRepo),
		NewPythonTool(),
		NewSkillTool("skills"),
		NewStudyPlanTool("data/study_plans"),
		NewFileSystemTool("data/workdir"),
		NewKnowledgeSearchTool(r),
		NewReportTool(reportRepo),
	}
}
