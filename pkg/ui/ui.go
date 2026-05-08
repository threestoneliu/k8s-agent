package ui

import (
	"time"

	"k8s-agent/pkg/session"
)

// UI 接口定义了与用户界面交互的抽象
// TUI、Web 等不同实现必须实现此接口
type UI interface {
	// SendMessage 发送一条消息给 UI 层展示
	SendMessage(msg *session.Message)

	// SendProgress 发送进度更新（工具调用开始、工具结果等）
	SendProgress(progress Progress)

	// Done 标记处理完成
	Done()

	// Error 发送错误信息
	Error(err error)

	// ClusterName 返回当前集群名称
	ClusterName() string

	// SetClusterName 设置当前集群名称
	SetClusterName(clusterName string)
}

// Progress 表示处理过程中的进度事件
type Progress struct {
	Type       string    // "text", "tool_call_start", "tool_result", "done"
	Content    string    // 文本内容或最终回复
	ToolName   string    // 工具名称
	ToolArgs   string    // 工具参数
	ToolResult string    // 工具执行结果
	ToolSuccess bool     // 工具是否执行成功
	Timestamp  time.Time // 时间戳
}

// Message 表示需要展示给用户的消息
type Message struct {
	Role      session.Role        // 角色：user, assistant
	Content   string              // 消息内容
	ToolCalls []session.ToolCall  // 工具调用列表
	Timestamp time.Time           // 时间戳
}
