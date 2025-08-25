package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DelGuard 安装参数结构体
type InstallOptions struct {
	Interactive bool   // 是否交互安装
	Overwrite   bool   // 是否覆盖已有别名
	Language    string // 语言代码（如 zh-CN, en-US）
	Silent      bool   // 是否静默安装
}

// 解析命令行参数，自动设置安装选项
func ParseInstallOptions() InstallOptions {
	opts := InstallOptions{
		Interactive: true,
		Overwrite:   false,
		Language:    "auto",
		Silent:      false,
	}
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--silent", "-s":
			opts.Silent = true
			opts.Interactive = false
		case "--interactive", "-i":
			opts.Interactive = true
			opts.Silent = false
		case "--overwrite", "-f":
			opts.Overwrite = true
		case "--lang":
			if i+1 < len(args) {
				opts.Language = args[i+1]
				i++
			}
		default:
			if strings.HasPrefix(arg, "--lang=") {
				opts.Language = strings.TrimPrefix(arg, "--lang=")
			}
		}
	}
	return opts
}

// PowerShellVersion PowerShell版本信息
type PowerShellVersion struct {
	Name        string
	Command     string
	ProfilePath string
	Version     string
	Available   bool
}

// installUnixAliases 是对 installUnixShellAliases 的包装，兼容旧调用点
func installUnixAliases(defaultInteractive bool, overwrite bool) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}
	return installUnixShellAliases(exePath, defaultInteractive, overwrite)
}

// 卸载别名统一入口
func uninstallAliases() error {
	switch runtime.GOOS {
	case "windows":
		return uninstallWindowsAliases()
	case "darwin", "linux":
		return uninstallUnixAliases()
	default:
		return ErrUnsupportedPlatform
	}
}

// Windows 卸载：PowerShell + CMD
func uninstallWindowsAliases() error {
	var psErr, cmdErr error

	if err := uninstallPowerShellAliases(); err != nil {
		psErr = err
		log.Printf("[ERROR] PowerShell别名卸载失败: %s", err.Error())
	} else {
		log.Println("[INFO] PowerShell别名已卸载")
	}

	if err := uninstallCmdAliases(); err != nil {
		cmdErr = err
		log.Printf("[ERROR] CMD别名卸载失败: %s", err.Error())
	} else {
		log.Println("[INFO] CMD别名已卸载")
	}

	if psErr != nil && cmdErr != nil {
		log.Printf("[FATAL] Windows别名卸载失败: PowerShell=%v; CMD=%v", psErr, cmdErr)
		return fmt.Errorf("Windows 别名卸载失败: PowerShell=%v; CMD=%v", psErr, cmdErr)
	}
	return nil
}

// 卸载 PowerShell 别名：清理各版本 Profile 中的 DelGuard 配置块
func uninstallPowerShellAliases() error {
	versions := detectPowerShellVersions()
	if len(versions) == 0 {
		return fmt.Errorf("未检测到可用的PowerShell版本")
	}

	var errs []string
	for _, v := range versions {
		if !v.Available || v.ProfilePath == "" {
			continue
		}
		content := ""
		if b, err := os.ReadFile(v.ProfilePath); err == nil {
			content = string(b)
		} else {
			continue
		}
		cleaned := removeOldDelGuardConfig(content)
		if cleaned == content {
			continue
		}
		if err := os.WriteFile(v.ProfilePath, []byte(cleaned), 0o644); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", v.Name, err))
			continue
		}
		log.Printf("[INFO] 已从 %s 移除 DelGuard 配置: %s", v.Name, v.ProfilePath)
	}

	if len(errs) > 0 {
		log.Printf("[WARN] 部分PowerShell配置卸载失败: %s", strings.Join(errs, "; "))
		return fmt.Errorf("部分PowerShell配置卸载失败: %s", strings.Join(errs, "; "))
	}
	return nil
}

// 卸载 CMD 别名：移除 AutoRun 中的宏文件引用并删除宏文件
func uninstallCmdAliases() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户主目录失败: %w", err)
	}
	macroPath := filepath.Join(homeDir, "delguard_macros.cmd")

	// 读取现有 AutoRun 设置
	key := `HKCU\Software\Microsoft\Command Processor`
	existing := ""
	out, err := exec.Command("reg", "query", key, "/v", "AutoRun").CombinedOutput()
	if err == nil {
		text := string(out)
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

	// 从 existing 中移除我们的宏命令
	if existing != "" {
		cleaned := removeCmdAutoRun(existing, macroPath)
		if cleaned != existing {
			// 写回注册表（若为空则清空）
			if strings.TrimSpace(cleaned) == "" {
				// 设置为空字符串
				if err := exec.Command("reg", "add", key, "/v", "AutoRun", "/t", "REG_SZ", "/d", "", "/f").Run(); err != nil {
					fmt.Printf(T("⚠️  清空AutoRun失败: %s\n"), err.Error())
				}
			} else {
				if err := exec.Command("reg", "add", key, "/v", "AutoRun", "/t", "REG_SZ", "/d", cleaned, "/f").Run(); err != nil {
					fmt.Printf(T("⚠️  更新AutoRun失败: %s\n"), err.Error())
				}
			}
		}
	}

	// 删除宏文件（若存在）
	if _, err := os.Stat(macroPath); err == nil {
		_ = os.Remove(macroPath)
	}

	return nil
}

// 移除 AutoRun 中的宏文件命令片段
func removeCmdAutoRun(existing, macroPath string) string {
	macroCmd := fmt.Sprintf(`doskey /macrofile="%s"`, macroPath)
	// 情况1：单独只有宏命令
	if strings.TrimSpace(existing) == macroCmd {
		return ""
	}
	// 情况2：以 & 连接的多命令，移除其中包含宏命令的部分
	parts := strings.Split(existing, "&")
	var kept []string
	for _, p := range parts {
		if !strings.Contains(p, "/macrofile=") || !strings.Contains(p, macroPath) {
			kept = append(kept, strings.TrimSpace(p))
		}
	}
	return strings.TrimSpace(strings.Join(kept, " & "))
}

// Unix 卸载：移除各 shell 配置文件中的别名行
func uninstallUnixAliases() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf(T("无法获取用户主目录: %w"), err)
	}
	shellConfigs := []string{".bashrc", ".bash_profile", ".zshrc", ".profile"}
	changed := false
	for _, cfg := range shellConfigs {
		p := filepath.Join(homeDir, cfg)
		if err := removeUnixAliasesFromShellConfig(p); err == nil {
			changed = true
			fmt.Printf(T("已从 %s 移除 DelGuard 别名\n"), p)
		}
	}
	if !changed {
		return fmt.Errorf("未在常见 shell 配置中发现 DelGuard 别名")
	}
	return nil
}

// 移除 Unix shell 配置中的 DelGuard 别名
func removeUnixAliasesFromShellConfig(configPath string) error {
	b, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(b), "\n")
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "DelGuard 别名") ||
			strings.HasPrefix(trimmed, "alias del='") ||
			strings.HasPrefix(trimmed, "alias rm='") ||
			strings.HasPrefix(trimmed, "alias cp='") {
			// 跳过这些行
			continue
		}
		out = append(out, line)
	}
	return os.WriteFile(configPath, []byte(strings.Join(out, "\n")), 0o644)
}

// detectPowerShellVersions 检测系统中所有可用的PowerShell版本
func detectPowerShellVersions() []PowerShellVersion {
	var versions []PowerShellVersion

	// PowerShell 7+ (pwsh)
	if pwshPath, err := exec.LookPath("pwsh"); err == nil {
		// 获取版本信息
		cmd := exec.Command(pwshPath, "-Command", "$PSVersionTable.PSVersion.ToString()")
		if output, err := cmd.Output(); err == nil {
			version := strings.TrimSpace(string(output))
			profilePath := getUserProfilePath(pwshPath, "pwsh")
			versions = append(versions, PowerShellVersion{
				Name:        "PowerShell 7+",
				Command:     pwshPath,
				ProfilePath: profilePath,
				Version:     version,
				Available:   true,
			})
		}
	}

	// Windows PowerShell 5.1 (powershell)
	if psPath, err := exec.LookPath("powershell"); err == nil {
		// 获取版本信息
		cmd := exec.Command(psPath, "-Command", "$PSVersionTable.PSVersion.ToString()")
		if output, err := cmd.Output(); err == nil {
			version := strings.TrimSpace(string(output))
			profilePath := getUserProfilePath(psPath, "powershell")
			versions = append(versions, PowerShellVersion{
				Name:        "Windows PowerShell",
				Command:     psPath,
				ProfilePath: profilePath,
				Version:     version,
				Available:   true,
			})
		}
	}

	return versions
}

// getUserProfilePath 获取指定PowerShell版本的用户配置文件路径
func getUserProfilePath(psCommand, psType string) string {
	cmd := exec.Command(psCommand, "-Command", "Write-Output $PROFILE")
	if output, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(output))
	}

	// 如果无法获取，使用默认路径
	homeDir, _ := os.UserHomeDir()
	switch psType {
	case "pwsh":
		return filepath.Join(homeDir, ".config", "powershell", "Microsoft.PowerShell_profile.ps1")
	case "powershell":
		return filepath.Join(homeDir, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
	default:
		return filepath.Join(homeDir, ".config", "powershell", "Microsoft.PowerShell_profile.ps1")
	}
}

// InstallStep 安装步骤
type InstallStep struct {
	Name        string
	Execute     func() error
	Rollback    func() error
	Description string
}

// InstallTransaction 事务性安装管理器
type InstallTransaction struct {
	Steps     []InstallStep
	Completed []int
	FailedAt  int
	Err       error
}

// NewInstallTransaction 创建新的安装事务
func NewInstallTransaction() *InstallTransaction {
	return &InstallTransaction{
		Steps:     make([]InstallStep, 0),
		Completed: make([]int, 0),
	}
}

// AddStep 添加安装步骤
func (t *InstallTransaction) AddStep(name string, execute, rollback func() error, description string) {
	t.Steps = append(t.Steps, InstallStep{
		Name:        name,
		Execute:     execute,
		Rollback:    rollback,
		Description: description,
	})
}

// Execute 执行安装事务
func (t *InstallTransaction) Execute() error {
	fmt.Println(T("=== 开始事务性安装 ==="))

	for i, step := range t.Steps {
		fmt.Printf(T("\n[%d/%d] 执行: %s\n"), i+1, len(t.Steps), step.Description)

		if err := step.Execute(); err != nil {
			fmt.Printf(T("   ❌ 安装步骤失败: %s\n"), err.Error())
			t.FailedAt = i
			t.Err = err

			// 执行回滚
			t.rollback()
			return err
		}

		fmt.Printf(T("   ✅ 步骤完成: %s\n"), step.Name)
		t.Completed = append(t.Completed, i)
	}

	fmt.Println(T("\n=== 安装成功完成 ==="))
	return nil
}

// rollback 执行回滚操作
func (t *InstallTransaction) rollback() {
	fmt.Println(T("\n=== 开始回滚操作 ==="))

	// 按相反顺序执行回滚
	for i := len(t.Completed) - 1; i >= 0; i-- {
		stepIndex := t.Completed[i]
		step := t.Steps[stepIndex]

		fmt.Printf(T("   🔁 回滚步骤: %s\n"), step.Name)
		if err := step.Rollback(); err != nil {
			fmt.Printf(T("      警告: 回滚失败: %s\n"), err.Error())
		} else {
			fmt.Printf(T("      ✅ 回滚成功: %s\n"), step.Name)
		}
	}

	fmt.Println(T("=== 回滚操作完成 ==="))
}

// installAliases 安装shell别名（Windows: CMD + PowerShell；Unix: bash/zsh 等）
// defaultInteractive: 是否将 del/rm 默认指向 delguard -i（交互删除）
func installAliases(defaultInteractive bool, overwrite bool) error {
	switch runtime.GOOS {
	case "windows":
		return installWindowsAliases(defaultInteractive, overwrite)
	case "darwin":
		return installUnixAliases(defaultInteractive, overwrite)
	case "linux":
		return installUnixAliases(defaultInteractive, overwrite)
	default:
		return ErrUnsupportedPlatform
	}
}

// Windows: 智能安装 PowerShell 和 CMD 的别名，支持多版本并提供详细反馈
func installWindowsAliases(defaultInteractive bool, overwrite bool) error {
	fmt.Println(T("=== 开始 Windows 别名安装 ==="))

	var psOK, cmdOK bool
	var psErr, cmdErr error

	// 安装PowerShell别名
	fmt.Println(T("\n1. 安装PowerShell别名..."))
	if err := installPowerShellAliases(defaultInteractive, overwrite); err != nil {
		psErr = err
		fmt.Printf(T("   ❌ PowerShell别名安装失败: %s\n"), err.Error())
	} else {
		psOK = true
		fmt.Println(T("   ✅ PowerShell别名安装成功"))
	}

	// 安装CMD别名
	fmt.Println(T("\n2. 安装CMD别名..."))
	if err := installCmdAliases(defaultInteractive, overwrite); err != nil {
		cmdErr = err
		fmt.Printf(T("   ❌ CMD别名安装失败: %s\n"), err.Error())
	} else {
		cmdOK = true
		fmt.Println(T("   ✅ CMD别名安装成功"))
	}

	// 总结安装结果
	fmt.Println(T("\n=== 安装结果总结 ==="))
	if psOK && cmdOK {
		fmt.Println(T("✅ 所有别名安装成功"))
		fmt.Println(T("📋 生效方式:"))
		fmt.Println(T("   PowerShell: 重启PowerShell 或执行 . $PROFILE"))
		fmt.Println(T("   CMD: 新开一个CMD窗口"))
		return nil
	} else if psOK || cmdOK {
		fmt.Println(T("⚠️  部分别名安装成功"))
		if psOK {
			fmt.Println(T("✅ PowerShell别名可用"))
		}
		if cmdOK {
			fmt.Println(T("✅ CMD别名可用"))
		}
		if psErr != nil {
			fmt.Printf(T("❌ PowerShell问题: %s\n"), psErr.Error())
		}
		if cmdErr != nil {
			fmt.Printf(T("❌ CMD问题: %s\n"), cmdErr.Error())
		}
		return nil
	} else {
		fmt.Println(T("❌ 所有别名安装失败"))
		var errors []string
		if psErr != nil {
			errors = append(errors, fmt.Sprintf("PowerShell: %s", psErr.Error()))
		}
		if cmdErr != nil {
			errors = append(errors, fmt.Sprintf("CMD: %s", cmdErr.Error()))
		}
		return fmt.Errorf("Windows别名安装完全失败:\n%s", strings.Join(errors, "\n"))
	}
}

// PowerShell 别名：智能检测并安装到所有可用的PowerShell版本
func installPowerShellAliases(defaultInteractive bool, overwrite bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户主目录失败: %w", err)
	}

	// 检测所有可用的PowerShell版本和配置文件路径
	versions := []PowerShellVersion{
		{Name: "PowerShell 7+", Command: "pwsh"},
		{Name: "Windows PowerShell 5.1", Command: "powershell"},
	}

	// 检测每个PowerShell版本
	var availableVersions []PowerShellVersion
	for _, version := range versions {
		// 检查版本可用性
		cmd := exec.Command(version.Command, "-NoProfile", "-Command", "$PSVersionTable.PSVersion.Major")
		output, err := cmd.Output()
		if err != nil {
			continue // 该版本不可用
		}
		version.Version = strings.TrimSpace(string(output))
		version.Available = true

		// 获取Profile路径
		cmd = exec.Command(version.Command, "-NoProfile", "-Command", "$PROFILE")
		profileOutput, err := cmd.Output()
		if err == nil {
			version.ProfilePath = strings.TrimSpace(string(profileOutput))
		} else {
			// 回退到默认路径
			if version.Command == "pwsh" {
				version.ProfilePath = filepath.Join(homeDir, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
			} else {
				version.ProfilePath = filepath.Join(homeDir, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
			}
		}

		availableVersions = append(availableVersions, version)
	}

	if len(availableVersions) == 0 {
		return fmt.Errorf("未检测到可用的PowerShell版本")
	}

	// 显示检测结果
	fmt.Println(T("检测到的PowerShell版本:"))
	for _, version := range availableVersions {
		fmt.Printf(T("  %s (版本 %s): %s\n"), version.Name, version.Version, version.ProfilePath)
	}

	// 为每个版本安装别名
	var installErrors []string
	successCount := 0

	for _, version := range availableVersions {
		if err := installToSinglePowerShell(version, defaultInteractive, overwrite); err != nil {
			installErrors = append(installErrors, fmt.Sprintf("%s: %v", version.Name, err))
		} else {
			successCount++
		}
	}

	// 报告安装结果
	if successCount > 0 {
		fmt.Printf(T("成功安装到 %d/%d 个PowerShell版本\n"), successCount, len(availableVersions))
		if len(installErrors) > 0 {
			fmt.Println(T("部分安装失败:"))
			for _, errMsg := range installErrors {
				fmt.Printf(T("  - %s\n"), errMsg)
			}
		}
		return nil
	} else {
		return fmt.Errorf("所有PowerShell版本安装失败: %s", strings.Join(installErrors, "; "))
	}
}

// installToSinglePowerShell 安装别名到单个PowerShell版本
func installToSinglePowerShell(version PowerShellVersion, defaultInteractive bool, overwrite bool) error {
	profilePath := version.ProfilePath

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(profilePath), 0o755); err != nil {
		return fmt.Errorf("创建PowerShell配置目录失败: %w", err)
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	// 生成稳健的PowerShell别名配置，支持所有5个命令
	aliasContent := generateRobustPowerShellConfig(exePath, version.Name, defaultInteractive)

	// 智能处理现有配置文件
	content := ""
	if b, err := os.ReadFile(profilePath); err == nil {
		content = string(b)
		// 移除旧的DelGuard配置块（支持多种格式）
		content = removeOldDelGuardConfig(content)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("读取PowerShell配置文件失败: %w", err)
	}

	// 检查是否已存在相同配置
	if strings.Contains(content, "DelGuard PowerShell Configuration") {
		if !overwrite {
			fmt.Printf(T("  %s: 别名已存在，跳过安装\n"), version.Name)
			return nil
		}
		fmt.Printf(T("  %s: 已覆盖原有别名配置\n"), version.Name)
	}

	// 添加新的别名配置
	content = strings.TrimRight(content, "\n") + "\n" + aliasContent + "\n"

	// 验证生成的配置语法
	if err := validatePowerShellSyntax(aliasContent); err != nil {
		return fmt.Errorf("PowerShell配置语法验证失败: %w", err)
	}

	// 写入配置文件
	if err := os.WriteFile(profilePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("写入PowerShell配置文件失败: %w", err)
	}

	fmt.Printf(T("  %s: 已安装到 %s\n"), version.Name, profilePath)
	return nil
}

// generateRobustPowerShellConfig 生成稳健的PowerShell配置，支持所有5个命令
func generateRobustPowerShellConfig(exePath, versionName string, defaultInteractive bool) string {
	// 使用单引号避免路径转义问题
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	interactiveFlag := ""
	if defaultInteractive {
		interactiveFlag = " -i"
	}

	config := fmt.Sprintf(`
# DelGuard PowerShell Configuration
# Generated: %s
# Version: DelGuard 1.0 for %s
# Supports: del, rm, cp, copy, delguard commands

$delguardPath = '%s'

if (Test-Path $delguardPath) {
    # Remove existing aliases to prevent conflicts
    try {
        Remove-Item Alias:del -Force -ErrorAction SilentlyContinue
        Remove-Item Alias:rm -Force -ErrorAction SilentlyContinue
        Remove-Item Alias:cp -Force -ErrorAction SilentlyContinue
        Remove-Item Alias:copy -Force -ErrorAction SilentlyContinue
    } catch { }
    
    # Define robust alias functions for all 5 commands
    function global:del {
        param([Parameter(ValueFromRemainingArguments)]$Arguments)
        & $delguardPath%s $Arguments
    }
    
    function global:rm {
        param([Parameter(ValueFromRemainingArguments)]$Arguments)
        & $delguardPath%s $Arguments
    }
    
    function global:cp {
        param([Parameter(ValueFromRemainingArguments)]$Arguments)
        & $delguardPath --cp $Arguments
    }
    
    function global:copy {
        param([Parameter(ValueFromRemainingArguments)]$Arguments)
        & $delguardPath --cp $Arguments
    }
    
    function global:delguard {
        param([Parameter(ValueFromRemainingArguments)]$Arguments)
        & $delguardPath $Arguments
    }
    
    # Show loading message only once per session
    if (-not $global:DelGuardLoaded) {
        Write-Host 'DelGuard aliases loaded successfully' -ForegroundColor Green
        Write-Host 'Commands: del, rm, cp, copy, delguard' -ForegroundColor Cyan
        Write-Host 'Use --help for detailed help' -ForegroundColor Gray
        $global:DelGuardLoaded = $true
    }
} else {
    Write-Warning "DelGuard executable not found: $delguardPath"
}
# End DelGuard Configuration
`, timestamp, versionName, exePath, interactiveFlag, interactiveFlag)

	return config
}

// validatePowerShellSyntax 验证PowerShell配置语法
func validatePowerShellSyntax(config string) error {
	// 基本语法检查
	lines := strings.Split(config, "\n")
	braceCount := 0
	parenCount := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// 检查括号匹配
		for _, char := range line {
			switch char {
			case '{':
				braceCount++
			case '}':
				braceCount--
			case '(':
				parenCount++
			case ')':
				parenCount--
			}
		}

		// 检查常见语法错误
		if strings.Contains(line, "if (-not )") {
			return fmt.Errorf("第%d行语法错误: if条件为空", i+1)
		}

		if strings.Contains(line, " = True") && !strings.Contains(line, "$true") {
			return fmt.Errorf("第%d行语法错误: 应使用$true而不是True", i+1)
		}
	}

	if braceCount != 0 {
		return fmt.Errorf("大括号不匹配: %d", braceCount)
	}

	if parenCount != 0 {
		return fmt.Errorf("小括号不匹配: %d", parenCount)
	}

	return nil
}

// removeOldDelGuardConfig 智能移除旧的DelGuard配置
func removeOldDelGuardConfig(content string) string {
	lines := strings.Split(content, "\n")
	var newLines []string
	skip := false
	delGuardBlockFound := false

	for _, line := range lines {
		// 检测DelGuard配置开始（支持多种格式）
		if strings.Contains(line, "DelGuard Safe Delete Aliases") ||
			strings.Contains(line, "DelGuard 安全删除别名") ||
			strings.Contains(line, "# DelGuard") {
			skip = true
			delGuardBlockFound = true
			continue
		}

		// 如果在跳过模式中
		if skip {
			// 检测配置块结束的多种情况
			trimmedLine := strings.TrimSpace(line)

			// 空行或注释行继续跳过
			if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
				continue
			}

			// DelGuard相关行继续跳过
			if strings.Contains(strings.ToLower(line), "delguard") ||
				strings.Contains(line, "Write-Host") ||
				strings.Contains(line, "function global:") ||
				strings.Contains(line, "Remove-Item") ||
				strings.Contains(line, "try {") ||
				strings.Contains(line, "} catch") ||
				strings.Contains(line, "$env:DELGUARD_LOADED") {
				continue
			}

			// 遇到其他有效内容，结束跳过
			skip = false
			newLines = append(newLines, line)
		} else {
			newLines = append(newLines, line)
		}
	}

	// 清理结果
	result := strings.Join(newLines, "\n")
	if delGuardBlockFound {
		// 清理多余的空行
		result = strings.TrimRight(result, "\n") + "\n"
	}

	return result
}

// CMD 别名：创建 doskey 宏文件并设置 AutoRun 加载，增强安全性和错误处理
func installCmdAliases(defaultInteractive bool, overwrite bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户主目录失败: %w", err)
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	// 宏文件路径
	macroPath := filepath.Join(homeDir, "delguard_macros.cmd")

	di := ""
	if defaultInteractive {
		di = "-i "
	}

	// 创建更健壮的宏文件，包含错误处理和版本信息
	macroContent := fmt.Sprintf(`@echo off
rem DelGuard CMD 别名宏文件
rem Generated: %s
rem Version: DelGuard 1.0
rem 使用更健壮的命令调用

rem 检查DelGuard可执行文件是否存在
if not exist "%s" (
    echo 错误: DelGuard 可执行文件不存在: %s
    echo 请检查安装或重新安装 DelGuard
    exit /b 1
)

rem 定义别名宏
doskey del="%s" %s$*
doskey rm="%s" %s$*
doskey cp="%s" --cp $*
doskey delguard="%s" $*

rem 显示成功加载信息（仅显示一次）
if not defined DELGUARD_CMD_LOADED (
    echo DelGuard CMD 别名已加载
    set DELGUARD_CMD_LOADED=1
)
`, time.Now().Format(TimeFormatStandard), exePath, exePath, exePath, di, exePath, di, exePath, exePath)

	// 智能处理现有宏文件
	if b, err := os.ReadFile(macroPath); err == nil {
		content := string(b)
		if strings.Contains(content, "DelGuard CMD 别名宏文件") {
			if !overwrite {
				fmt.Printf(T("   已存在CMD别名宏文件，跳过覆盖: %s\n"), macroPath)
				return updateCmdAutoRun(macroPath)
			}
			// 覆盖现有文件
			if err := os.WriteFile(macroPath, []byte(macroContent), 0o644); err != nil {
				return fmt.Errorf("更新宏文件失败: %w", err)
			}
			fmt.Printf(T("   已覆盖CMD别名宏文件: %s\n"), macroPath)
			return updateCmdAutoRun(macroPath)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("检查宏文件失败: %w", err)
	}

	// 创建新的宏文件
	if err := os.WriteFile(macroPath, []byte(macroContent), 0o644); err != nil {
		return fmt.Errorf("写入宏文件失败: %w", err)
	}

	fmt.Printf(T("   已创建CMD别名宏文件: %s\n"), macroPath)
	return updateCmdAutoRun(macroPath)
}

// updateCmdAutoRun 更新CMD AutoRun注册表设置
func updateCmdAutoRun(macroPath string) error {
	// Windows注册表键路径
	key := `HKCU\Software\Microsoft\Command Processor`

	// 读取现有 AutoRun 设置
	existing := ""
	out, err := exec.Command("reg", "query", key, "/v", "AutoRun").CombinedOutput()
	if err != nil {
		// 如果键不存在，记录信息但继续
		fmt.Printf(T("   检测到AutoRun键不存在，将创建新键\n"))
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

	// 构建新的AutoRun命令
	macroCmd := fmt.Sprintf(`doskey /macrofile="%s"`, macroPath)
	newVal := macroCmd

	if existing != "" {
		// 检查是否已经包含我们的宏文件
		if strings.Contains(existing, macroPath) {
			fmt.Printf(T("   AutoRun中已包含我们的宏文件，无需更新\n"))
			return nil
		}
		// 保留原有 AutoRun 并添加我们的
		newVal = existing + " & " + macroCmd
	}

	// 写入 AutoRun设置
	cmd := exec.Command("reg", "add", key, "/v", "AutoRun", "/t", "REG_SZ", "/d", newVal, "/f")
	if err := cmd.Run(); err != nil {
		// 如果注册表操作失败，提供备用方案
		fmt.Printf(T("   ⚠️  设置AutoRun失败: %s\n"), err.Error())
		fmt.Printf(T("   📋 手动启用方法: 在CMD中执行\n"))
		fmt.Printf(T("      doskey /macrofile=\"%s\"\n"), macroPath)
		fmt.Printf(T("   或者使用管理员权限重新运行安装\n"))
		return nil // 不返回错误，只是警告
	}

	fmt.Printf(T("   ✅ 已更新CMD AutoRun设置\n"))
	return nil
}

// installUnixShellAliases 为Unix shell安装别名，支持bash、zsh、fish、PowerShell for Linux等
func installUnixShellAliases(exePath string, defaultInteractive bool, overwrite bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf(T("无法获取用户主目录: %w"), err)
	}

	fmt.Println(T("=== Unix/Linux Shell 别名安装 ==="))

	// 检测并安装到各种shell
	var installResults []string

	// 1. Bash 支持
	if err := installBashAliases(homeDir, exePath, defaultInteractive, overwrite); err == nil {
		installResults = append(installResults, "✅ Bash")
	} else {
		installResults = append(installResults, fmt.Sprintf("❌ Bash: %v", err))
	}

	// 2. Zsh 支持
	if err := installZshAliases(homeDir, exePath, defaultInteractive, overwrite); err == nil {
		installResults = append(installResults, "✅ Zsh")
	} else {
		installResults = append(installResults, fmt.Sprintf("❌ Zsh: %v", err))
	}

	// 3. Fish 支持
	if err := installFishAliases(homeDir, exePath, defaultInteractive, overwrite); err == nil {
		installResults = append(installResults, "✅ Fish")
	} else {
		installResults = append(installResults, fmt.Sprintf("❌ Fish: %v", err))
	}

	// 4. PowerShell for Linux 支持
	if err := installPowerShellLinuxAliases(homeDir, exePath, defaultInteractive, overwrite); err == nil {
		installResults = append(installResults, "✅ PowerShell (Linux)")
	} else {
		installResults = append(installResults, fmt.Sprintf("❌ PowerShell (Linux): %v", err))
	}

	// 5. 通用 .profile 支持
	if err := installProfileAliases(homeDir, exePath, defaultInteractive, overwrite); err == nil {
		installResults = append(installResults, "✅ .profile")
	} else {
		installResults = append(installResults, fmt.Sprintf("❌ .profile: %v", err))
	}

	// 显示安装结果
	fmt.Println(T("\n=== 安装结果 ==="))
	successCount := 0
	for _, result := range installResults {
		fmt.Printf(T("  %s\n"), result)
		if strings.HasPrefix(result, "✅") {
			successCount++
		}
	}

	if successCount == 0 {
		return fmt.Errorf("所有shell配置安装失败")
	}

	fmt.Printf(T("\n✅ 成功安装到 %d 个shell环境\n"), successCount)
	fmt.Println(T("📋 生效方式: 重新打开终端或执行 source ~/.bashrc (或对应配置文件)"))

	return nil
}

// installBashAliases 安装Bash别名
func installBashAliases(homeDir, exePath string, defaultInteractive bool, overwrite bool) error {
	configs := []string{".bashrc", ".bash_profile"}
	installed := false

	for _, config := range configs {
		configPath := filepath.Join(homeDir, config)
		if _, err := os.Stat(configPath); err == nil {
			if err := appendAliasesToShellConfig(configPath, exePath, defaultInteractive, overwrite); err != nil {
				continue
			}
			fmt.Printf(T("  已更新 %s\n"), config)
			installed = true
		}
	}

	if !installed {
		// 创建 .bashrc
		bashrcPath := filepath.Join(homeDir, ".bashrc")
		if err := appendAliasesToShellConfig(bashrcPath, exePath, defaultInteractive, overwrite); err != nil {
			return err
		}
		fmt.Printf(T("  已创建 %s\n"), ".bashrc")
	}

	return nil
}

// installZshAliases 安装Zsh别名
func installZshAliases(homeDir, exePath string, defaultInteractive bool, overwrite bool) error {
	zshrcPath := filepath.Join(homeDir, ".zshrc")

	// 检查是否安装了zsh
	if _, err := exec.LookPath("zsh"); err != nil {
		return fmt.Errorf("zsh未安装")
	}

	if err := appendAliasesToShellConfig(zshrcPath, exePath, defaultInteractive, overwrite); err != nil {
		return err
	}

	fmt.Printf(T("  已更新 %s\n"), ".zshrc")
	return nil
}

// installFishAliases 安装Fish shell别名
func installFishAliases(homeDir, exePath string, defaultInteractive bool, overwrite bool) error {
	// 检查是否安装了fish
	if _, err := exec.LookPath("fish"); err != nil {
		return fmt.Errorf("fish未安装")
	}

	fishConfigDir := filepath.Join(homeDir, ".config", "fish")
	fishConfigPath := filepath.Join(fishConfigDir, "config.fish")

	// 确保目录存在
	if err := os.MkdirAll(fishConfigDir, 0o755); err != nil {
		return fmt.Errorf("创建fish配置目录失败: %w", err)
	}

	// Fish shell使用不同的别名语法
	interactiveFlag := ""
	if defaultInteractive {
		interactiveFlag = " -i"
	}

	fishAliases := fmt.Sprintf(`
# DelGuard Fish Shell 别名
# Generated: %s
alias del='%s%s'
alias rm='%s%s'
alias cp='%s --cp'
alias copy='%s --cp'
alias delguard='%s'
`, time.Now().Format("2006-01-02 15:04:05"), exePath, interactiveFlag, exePath, interactiveFlag, exePath, exePath, exePath)

	content := ""
	if b, err := os.ReadFile(fishConfigPath); err == nil {
		content = string(b)
		if strings.Contains(content, "# DelGuard Fish Shell 别名") && !overwrite {
			return fmt.Errorf("别名已存在")
		}
		// 移除旧的别名块
		content = removeDelGuardBlock(content, "# DelGuard Fish Shell 别名")
	}

	content = strings.TrimRight(content, "\n") + fishAliases + "\n"

	if err := os.WriteFile(fishConfigPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("写入fish配置失败: %w", err)
	}

	fmt.Printf(T("  已更新 %s\n"), "config.fish")
	return nil
}

// installPowerShellLinuxAliases 安装PowerShell for Linux别名
func installPowerShellLinuxAliases(homeDir, exePath string, defaultInteractive bool, overwrite bool) error {
	// 检查是否安装了PowerShell for Linux
	pwshPath, err := exec.LookPath("pwsh")
	if err != nil {
		return fmt.Errorf("PowerShell (pwsh)未安装")
	}

	// 获取PowerShell配置文件路径
	cmd := exec.Command(pwshPath, "-Command", "Write-Output $PROFILE")
	output, err := cmd.Output()
	if err != nil {
		// 使用默认路径
		profilePath := filepath.Join(homeDir, ".config", "powershell", "Microsoft.PowerShell_profile.ps1")
		return installPowerShellLinuxProfile(profilePath, exePath, defaultInteractive, overwrite)
	}

	profilePath := strings.TrimSpace(string(output))
	return installPowerShellLinuxProfile(profilePath, exePath, defaultInteractive, overwrite)
}

// installPowerShellLinuxProfile 安装PowerShell Linux配置文件
func installPowerShellLinuxProfile(profilePath, exePath string, defaultInteractive bool, overwrite bool) error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(profilePath), 0o755); err != nil {
		return fmt.Errorf("创建PowerShell配置目录失败: %w", err)
	}

	// 生成PowerShell配置
	aliasContent := generateRobustPowerShellConfig(exePath, "PowerShell Linux", defaultInteractive)

	content := ""
	if b, err := os.ReadFile(profilePath); err == nil {
		content = string(b)
		if strings.Contains(content, "DelGuard PowerShell Configuration") && !overwrite {
			return fmt.Errorf("别名已存在")
		}
		content = removeOldDelGuardConfig(content)
	}

	content = strings.TrimRight(content, "\n") + "\n" + aliasContent + "\n"

	if err := os.WriteFile(profilePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("写入PowerShell配置失败: %w", err)
	}

	fmt.Printf(T("  已更新 PowerShell配置: %s\n"), profilePath)
	return nil
}

// installProfileAliases 安装通用.profile别名
func installProfileAliases(homeDir, exePath string, defaultInteractive bool, overwrite bool) error {
	profilePath := filepath.Join(homeDir, ".profile")

	if err := appendAliasesToShellConfig(profilePath, exePath, defaultInteractive, overwrite); err != nil {
		return err
	}

	fmt.Printf(T("  已更新 %s\n"), ".profile")
	return nil
}

// removeDelGuardBlock 移除指定的DelGuard配置块
func removeDelGuardBlock(content, marker string) string {
	lines := strings.Split(content, "\n")
	var newLines []string
	skip := false

	for _, line := range lines {
		if strings.Contains(line, marker) {
			skip = true
			continue
		}

		if skip {
			// 检查是否到达块结束
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if !strings.HasPrefix(trimmed, "#") &&
				!strings.Contains(strings.ToLower(line), "delguard") &&
				!strings.Contains(line, "alias") {
				skip = false
				newLines = append(newLines, line)
			}
		} else {
			newLines = append(newLines, line)
		}
	}

	return strings.Join(newLines, "\n")
}

// appendAliasesToShellConfig 向shell配置文件追加别名
func appendAliasesToShellConfig(configPath, exePath string, defaultInteractive bool, overwrite bool) error {
	// 创建别名内容
	aliases := fmt.Sprintf(`
# DelGuard 别名
alias del='%s'
alias rm='%s'
alias cp='%s --cp'
`, exePath, exePath, exePath)

	if defaultInteractive {
		aliases = fmt.Sprintf(`
# DelGuard 别名 (交互模式)
alias del='%s -i'
alias rm='%s -i'
alias cp='%s --cp -i'
`, exePath, exePath, exePath)
	}

	content := ""
	if b, err := os.ReadFile(configPath); err == nil {
		content = string(b)
		if strings.Contains(content, "# DelGuard 别名") && !overwrite {
			fmt.Printf(T("%s 已存在 DelGuard 别名，跳过安装\n"), configPath)
			return nil
		}
		// 移除旧的别名块
		content = removeOldDelGuardConfig(content)
	}
	// 追加新别名
	content = strings.TrimRight(content, "\n") + aliases + "\n"
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf(T("写入配置文件失败: %w"), err)
	}
	return nil
}
