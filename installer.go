package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// installAliases 安装shell别名（Windows: CMD + PowerShell；Unix: bash/zsh 等）
// defaultInteractive: 是否将 del/rm 默认指向 delguard -i（交互删除）
func installAliases(defaultInteractive bool) error {
	switch runtime.GOOS {
	case "windows":
		return installWindowsAliases(defaultInteractive)
	case "darwin":
		return installUnixAliases(defaultInteractive)
	case "linux":
		return installUnixAliases(defaultInteractive)
	default:
		return ErrUnsupportedPlatform
	}
}

// Windows: 同时安装 PowerShell 和 CMD 的别名
func installWindowsAliases(defaultInteractive bool) error {
	var psOK, cmdOK bool
	if err := installPowerShellAliases(defaultInteractive); err != nil {
		fmt.Printf("安装PowerShell别名失败: %s\n", err.Error())
	} else {
		psOK = true
	}

	if err := installCmdAliases(defaultInteractive); err != nil {
		fmt.Printf("安装CMD别名失败: %s\n", err.Error())
	} else {
		cmdOK = true
	}

	if psOK || cmdOK {
		return nil
	}
	return fmt.Errorf("Windows别名安装失败（PowerShell与CMD均未成功）")
}

// PowerShell 别名：定义 del 函数并将 rm 指向 del
func installPowerShellAliases(defaultInteractive bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// 支持 PowerShell 7+ 和 Windows PowerShell 5.1 的不同路径
	psProfilePath := filepath.Join(homeDir, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	winPsProfilePath := filepath.Join(homeDir, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")

	// 优先使用 PowerShell 7+ 路径
	profilePath := psProfilePath
	if _, err := os.Stat(psProfilePath); os.IsNotExist(err) {
		// 如果 PowerShell 7+ 路径不存在，使用 Windows PowerShell 路径
		profilePath = winPsProfilePath
	}

	if err := os.MkdirAll(filepath.Dir(profilePath), 0o755); err != nil {
		return err
	}

	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	// 转义Windows路径中的反斜杠用于PowerShell
	escapedExePath := strings.ReplaceAll(exePath, "\\", "\\\\")

	di := ""
	if defaultInteractive {
		di = "-i "
	}

	aliasContent := fmt.Sprintf(`
# DelGuard 安全删除别名 (PowerShell)
# 移除内置别名以确保函数生效
if (Get-Command del -ErrorAction SilentlyContinue) {
    if (Get-Alias del -ErrorAction SilentlyContinue) { Remove-Item Alias:del -Force }
    if (Get-Command Remove-Item -ErrorAction SilentlyContinue) { Remove-Item Alias:del -Force -ErrorAction SilentlyContinue }
}
if (Get-Command rm -ErrorAction SilentlyContinue) {
    if (Get-Alias rm -ErrorAction SilentlyContinue) { Remove-Item Alias:rm -Force }
    if (Get-Command Remove-Item -ErrorAction SilentlyContinue) { Remove-Item Alias:rm -Force -ErrorAction SilentlyContinue }
}
if (Get-Command cp -ErrorAction SilentlyContinue) {
    if (Get-Alias cp -ErrorAction SilentlyContinue) { Remove-Item Alias:cp -Force }
    if (Get-Command Copy-Item -ErrorAction SilentlyContinue) { Remove-Item Alias:cp -Force -ErrorAction SilentlyContinue }
}
function global:del { 
    param([Parameter(ValueFromRemainingArguments)]$Arguments)
    & "%s" %s$Arguments 
}
function global:rm { 
    param([Parameter(ValueFromRemainingArguments)]$Arguments)
    & "%s" %s$Arguments 
}
function global:cp { 
    param([Parameter(ValueFromRemainingArguments)]$Arguments)
    & "%s" --cp $Arguments 
}
Set-Alias -Name del -Value del -Scope Global -Force
Set-Alias -Name rm -Value rm -Scope Global -Force
Set-Alias -Name cp -Value cp -Scope Global -Force
`, escapedExePath, di, escapedExePath, di, escapedExePath)

	// 检查并移除旧的别名配置
	content := ""
	if b, err := os.ReadFile(profilePath); err == nil {
		content = string(b)
		// 移除旧的配置
		lines := strings.Split(content, "\n")
		var newLines []string
		skip := false
		for _, line := range lines {
			if strings.Contains(line, "DelGuard 安全删除别名") {
				skip = true
				continue
			}
			if skip && strings.TrimSpace(line) == "" {
				skip = false
				continue
			}
			if !skip {
				newLines = append(newLines, line)
			}
		}
		content = strings.Join(newLines, "\n")
	}

	// 添加新的别名配置
	if !strings.Contains(content, "DelGuard 安全删除别名 (PowerShell)") {
		content += aliasContent
	}

	// 写入配置文件
	err = os.WriteFile(profilePath, []byte(content), 0o644)
	if err != nil {
		return err
	}

	fmt.Printf("已安装PowerShell别名到: %s\n", profilePath)
	fmt.Println("请重启PowerShell，或执行: . $PROFILE 以生效")
	return nil
}

// CMD 别名：创建 doskey 宏文件并设置 AutoRun 加载
func installCmdAliases(defaultInteractive bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	// 宏文件路径
	macroPath := filepath.Join(homeDir, "delguard_macros.cmd")

	di := ""
	if defaultInteractive {
		di = "-i "
	}

	// 创建更健壮的宏文件，包含错误处理
	macroContent := fmt.Sprintf(`@echo off
rem DelGuard CMD 别名宏文件
rem 使用更健壮的命令调用
if "%%1"=="" (
    "%s" %s%%*
) else (
    "%s" %s%%*
)
exit /b

rem 定义别名
doskey del="%s" %s$*
doskey rm="%s" %s$*
doskey cp="%s" --cp $*
doskey delguard="%s" $*
`, exePath, di, exePath, di, exePath, di, exePath, di, exePath, exePath)

	// 检查现有宏文件，避免重复
	if b, err := os.ReadFile(macroPath); err == nil {
		content := string(b)
		if strings.Contains(content, "DelGuard CMD 别名宏文件") {
			// 更新现有文件
			if err := os.WriteFile(macroPath, []byte(macroContent), 0o644); err != nil {
				return fmt.Errorf("更新宏文件失败: %w", err)
			}
			fmt.Printf("已更新CMD别名宏文件: %s\n", macroPath)
			fmt.Println("请新开一个 CMD 窗口以生效")
			return nil
		}
	}

	if err := os.WriteFile(macroPath, []byte(macroContent), 0o644); err != nil {
		return fmt.Errorf("写入宏文件失败: %w", err)
	}

	// 读取现有 AutoRun
	key := `HKCU\Software\Microsoft\Command Processor`
	existing := ""
	out, err := exec.Command("reg", "query", key, "/v", "AutoRun").CombinedOutput()
	if err != nil {
		// 如果键不存在，创建它
		existing = ""
	} else {
		text := string(out)
		// 解析 AutoRun 行
		for _, line := range strings.Split(text, "\n") {
			if strings.Contains(line, "AutoRun") && strings.Contains(line, "REG_SZ") {
				parts := strings.Split(line, "REG_SZ")
				if len(parts) > 1 {
					existing = strings.TrimSpace(parts[1])
				}
				break
			}
		}
	}

	macroCmd := fmt.Sprintf(`doskey /macrofile="%s"`, macroPath)
	newVal := macroCmd
	if existing != "" && !strings.Contains(existing, macroPath) {
		// 保留原有 AutoRun
		newVal = existing + " & " + macroCmd
	} else if existing != "" {
		newVal = existing
	}

	// 写入 AutoRun
	cmd := exec.Command("reg", "add", key, "/v", "AutoRun", "/t", "REG_SZ", "/d", newVal, "/f")
	if err := cmd.Run(); err != nil {
		// 如果注册表操作失败，提供备用方案
		fmt.Printf("警告：设置CMD AutoRun失败: %s\n", err.Error())
		fmt.Printf("您可以手动执行以下命令来启用CMD别名：\n")
		fmt.Printf("  在CMD中运行: doskey /macrofile=\"%s\"\n", macroPath)
		return nil
	}

	fmt.Printf("已安装CMD别名，宏文件: %s\n", macroPath)
	fmt.Println("请新开一个 CMD 窗口以生效")
	return nil
}

// Unix 系统：同时为 rm 和 del 设置别名
func installUnixAliases(defaultInteractive bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	di := ""
	if defaultInteractive {
		di = " -i"
	}

	aliasContent := fmt.Sprintf(`
# DelGuard 安全删除别名 (Unix)
alias rm='%s%s'
alias del='%s%s'
alias cp='%s --cp'
`, exePath, di, exePath, di, exePath)

	configFiles := []string{
		".bashrc",
		".zshrc",
		".profile",
	}

	installed := false
	for _, cfg := range configFiles {
		cfgPath := filepath.Join(homeDir, cfg)
		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			continue
		}
		if b, err := os.ReadFile(cfgPath); err == nil && strings.Contains(string(b), "DelGuard 安全删除别名 (Unix)") {
			fmt.Printf("别名已存在于: %s\n", cfgPath)
			installed = true
			continue
		}
		f, err := os.OpenFile(cfgPath, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			continue
		}
		if _, err := f.WriteString(aliasContent); err == nil {
			fmt.Printf("已安装别名到: %s\n", cfgPath)
			installed = true
		}
		_ = f.Close()
	}

	if !installed {
		return fmt.Errorf("未找到可写入的shell配置文件")
	}
	fmt.Println("请重启终端或执行: source ~/.bashrc (或对应文件) 以生效")
	return nil
}
