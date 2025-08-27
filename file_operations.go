package main

import (
	"delguard/utils"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// OverwriteProtector 覆盖保护器
type OverwriteProtector struct {
	Config *Config
}

// NewOverwriteProtector 创建新的覆盖保护器
func NewOverwriteProtector(config *Config) *OverwriteProtector {
	return &OverwriteProtector{
		Config: config,
	}
}

// ProtectOverwrite 保护文件不被意外覆盖
func (op *OverwriteProtector) ProtectOverwrite(filename string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filename); err == nil {
		// 文件存在，备份原文件到回收站
		if err := moveToTrashPlatform(filename); err != nil {
			return fmt.Errorf("备份文件到回收站失败: %v", err)
		}
	}
	return nil
}

// FileOperations 提供安全的文件操作
var FileOperations = &fileOperations{}

// SafeFileInfo 安全的文件信息结构
type SafeFileInfo struct {
	Path         string
	Size         int64
	Mode         os.FileMode
	ModTime      time.Time
	IsDir        bool
	Sha256       string
	RelativePath string
}

type fileOperations struct{}

// CopyFile 安全复制文件，支持覆盖保护
func (fo *fileOperations) CopyFile(source, destination string, protectOverwrite bool) error {
	if protectOverwrite {
		config, err := LoadConfig()
		if err != nil {
			log.Printf("[ERROR] 加载配置失败: %v", err)
			return fmt.Errorf("加载配置失败: %w", err)
		}
		if config.EnableOverwriteProtection {
			protector := NewOverwriteProtector(config)
			if err := protector.ProtectOverwrite(destination); err != nil {
				log.Printf("[ERROR] 覆盖保护失败: %v", err)
				return fmt.Errorf("覆盖保护失败: %w", err)
			}
		}
	}
	log.Printf("[INFO] 开始复制文件: %s -> %s", source, destination)
	return copyFile(source, destination)
}

// MoveFile 安全移动文件，支持覆盖保护
func (fo *fileOperations) MoveFile(source, destination string, protectOverwrite bool) error {
	if protectOverwrite {
		config, err := LoadConfig()
		if err != nil {
			log.Printf("[ERROR] 加载配置失败: %v", err)
			return fmt.Errorf("加载配置失败: %w", err)
		}
		if config.EnableOverwriteProtection {
			protector := NewOverwriteProtector(config)
			if err := protector.ProtectOverwrite(destination); err != nil {
				log.Printf("[ERROR] 覆盖保护失败: %v", err)
				return fmt.Errorf("覆盖保护失败: %w", err)
			}
		}
	}
	log.Printf("[INFO] 移动文件: %s -> %s", source, destination)
	return os.Rename(source, destination)
}

// WriteFile 安全写入文件，支持覆盖保护
func (fo *fileOperations) WriteFile(filename string, data []byte, perm os.FileMode, protectOverwrite bool) error {
	if protectOverwrite {
		config, err := LoadConfig()
		if err != nil {
			log.Printf("[ERROR] 加载配置失败: %v", err)
			return fmt.Errorf("加载配置失败: %w", err)
		}
		if config.EnableOverwriteProtection {
			protector := NewOverwriteProtector(config)
			if err := protector.ProtectOverwrite(filename); err != nil {
				log.Printf("[ERROR] 覆盖保护失败: %v", err)
				return fmt.Errorf("覆盖保护失败: %w", err)
			}
		}
	}
	log.Printf("[INFO] 写入文件: %s", filename)
	return os.WriteFile(filename, data, perm)
}

// CreateFile 安全创建文件，支持覆盖保护
func (fo *fileOperations) CreateFile(filename string, protectOverwrite bool) (*os.File, error) {
	if protectOverwrite {
		config, err := LoadConfig()
		if err != nil {
			log.Printf("[ERROR] 加载配置失败: %v", err)
			return nil, fmt.Errorf("加载配置失败: %w", err)
		}
		if config.EnableOverwriteProtection {
			protector := NewOverwriteProtector(config)
			if err := protector.ProtectOverwrite(filename); err != nil {
				log.Printf("[ERROR] 覆盖保护失败: %v", err)
				return nil, fmt.Errorf("覆盖保护失败: %w", err)
			}
		}
	}
	log.Printf("[INFO] 创建文件: %s", filename)
	return os.Create(filename)
}

// copyFile 实际执行文件复制操作
func copyFile(src, dst string) error {
	return utils.CopyFile(src, dst)
}

// BackupFileBeforeOverwrite 在覆盖前备份文件
func BackupFileBeforeOverwrite(filename string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil // 文件不存在，无需备份
	}

	// 生成备份文件名
	backupName := filename + ".backup." + time.Now().Format("20060102150405")

	// 复制文件到备份位置
	if err := copyFile(filename, backupName); err != nil {
		return fmt.Errorf("创建备份文件失败: %v", err)
	}

	// 将原文件移动到回收站
	if err := moveToTrashPlatform(filename); err != nil {
		// 如果移动到回收站失败，删除备份文件并返回错误
		os.Remove(backupName)
		return fmt.Errorf("移动原文件到回收站失败: %v", err)
	}

	return nil
}

// copyFileWithProgress 带进度显示的复制
func (fo *fileOperations) CopyFileWithProgress(source, destination string, protectOverwrite bool, progress func(bytesWritten, totalBytes int64)) error {
	if protectOverwrite {
		config, err := LoadConfig()
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		if config.EnableOverwriteProtection {
			protector := NewOverwriteProtector(config)
			if err := protector.ProtectOverwrite(destination); err != nil {
				return fmt.Errorf("覆盖保护失败: %w", err)
			}
		}
	}

	// 确保目标目录存在
	destDir := filepath.Dir(destination)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// 打开源文件
	srcFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// 获取文件信息
	info, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// 创建目标文件
	destFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// 执行带进度回调的复制
	buffer := make([]byte, 32*1024) // 32KB buffer
	total := info.Size()
	var written int64

	for {
		bytesRead, err := srcFile.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if bytesRead == 0 {
			break
		}

		_, writeErr := destFile.Write(buffer[:bytesRead])
		if writeErr != nil {
			return writeErr
		}

		written += int64(bytesRead)
		if progress != nil {
			progress(written, total)
		}
	}

	return destFile.Sync()
}

// SecureFileOperations 提供安全的文件操作
type SecureFileOperations struct{}

// CreateFile 创建文件，带安全检查
func (sfo *SecureFileOperations) CreateFile(name string) (*os.File, error) {
	// 启用覆盖保护
	if err := EnableOverwriteProtection(); err != nil {
		return nil, err
	}

	// 检查文件是否存在
	if _, err := os.Stat(name); err == nil {
		// 文件存在，备份原文件
		if err := BackupFileBeforeOverwrite(name); err != nil {
			return nil, fmt.Errorf("备份文件失败: %v", err)
		}
	}

	// 创建新文件
	file, err := os.Create(name)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// OpenFile 打开文件，带安全检查
func (sfo *SecureFileOperations) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	// 启用覆盖保护
	if flag&(os.O_TRUNC|os.O_WRONLY|os.O_RDWR) != 0 {
		if err := EnableOverwriteProtection(); err != nil {
			return nil, err
		}

		// 检查文件是否存在
		if _, err := os.Stat(name); err == nil {
			// 文件存在且将被修改，备份原文件
			if err := BackupFileBeforeOverwrite(name); err != nil {
				return nil, fmt.Errorf("备份文件失败: %v", err)
			}
		}
	}

	// 打开文件
	return os.OpenFile(name, flag, perm)
}

// CopyFile 复制文件，带安全检查
func (sfo *SecureFileOperations) CopyFile(src, dst string) error {
	// 启用覆盖保护
	if err := EnableOverwriteProtection(); err != nil {
		return err
	}

	// 检查目标文件是否存在
	if _, err := os.Stat(dst); err == nil {
		// 文件存在，备份原文件
		if err := BackupFileBeforeOverwrite(dst); err != nil {
			return fmt.Errorf("备份文件失败: %v", err)
		}
	}

	// 执行复制操作
	return copyFile(src, dst)
}

// WriteFile 写入文件，带安全检查
func (sfo *SecureFileOperations) WriteFile(name string, data []byte, perm os.FileMode) error {
	// 启用覆盖保护
	if err := EnableOverwriteProtection(); err != nil {
		return err
	}

	// 检查文件是否存在
	if _, err := os.Stat(name); err == nil {
		// 文件存在，备份原文件
		if err := BackupFileBeforeOverwrite(name); err != nil {
			return fmt.Errorf("备份文件失败: %v", err)
		}
	}

	// 写入文件
	return os.WriteFile(name, data, perm)
}

// Rename 重命名文件，带安全检查
func (sfo *SecureFileOperations) Rename(oldpath, newpath string) error {
	// 启用覆盖保护
	if err := EnableOverwriteProtection(); err != nil {
		return err
	}

	// 检查目标文件是否存在
	if _, err := os.Stat(newpath); err == nil {
		// 文件存在，备份原文件
		if err := BackupFileBeforeOverwrite(newpath); err != nil {
			return fmt.Errorf("备份文件失败: %v", err)
		}
	}

	// 创建覆盖保护器
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	protector := NewOverwriteProtector(config)
	if err := protector.ProtectOverwrite(newpath); err != nil {
		return fmt.Errorf("覆盖保护失败: %w", err)
	}

	// 重命名文件
	return os.Rename(oldpath, newpath)
}
