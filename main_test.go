package main

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// 设置测试环境
	InitGlobalLogger("DEBUG")

	// 运行测试
	code := m.Run()

	// 清理测试环境
	if logger != nil {
		logger.Close()
	}

	os.Exit(code)
}

func TestRunMainApp(t *testing.T) {
	tests := []struct {
		name         string
		setupContext func() (context.Context, context.CancelFunc)
		expectedExit int
		expectError  bool
	}{
		{
			name: "正常上下文",
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			expectedExit: ExitGeneralError, // 因为RunDelGuardApp还未实现
			expectError:  false,
		},
		{
			name: "已取消的上下文",
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // 立即取消
				return ctx, cancel
			},
			expectedExit: ExitUserCancelled,
			expectError:  false,
		},
		{
			name: "超时上下文",
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 1*time.Millisecond)
			},
			expectedExit: ExitUserCancelled,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.setupContext()
			defer cancel()

			// 对于已取消的上下文，稍等一下确保取消生效
			if tt.name == "已取消的上下文" || tt.name == "超时上下文" {
				time.Sleep(10 * time.Millisecond)
			}

			exitCode := runMainApp(ctx)

			if exitCode != tt.expectedExit {
				t.Errorf("runMainApp() = %v, want %v", exitCode, tt.expectedExit)
			}
		})
	}
}

func TestRunDelGuardApp(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		expectError bool
	}{
		{
			name:        "正常上下文",
			ctx:         context.Background(),
			expectError: true, // 当前实现返回错误
		},
		{
			name:        "已取消的上下文",
			ctx:         func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunDelGuardApp(tt.ctx)

			if (err != nil) != tt.expectError {
				t.Errorf("RunDelGuardApp() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// 基准测试
func BenchmarkRunMainApp(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runMainApp(ctx)
	}
}
