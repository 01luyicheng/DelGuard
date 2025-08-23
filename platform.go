package main

// GetUseTrash 返回是否使用回收站
func GetUseTrash() bool {
	// 从配置中获取值
	config, err := LoadConfig()
	if err != nil {
		// 出错时返回默认值
		return true
	}
	return config.UseRecycleBin
}

// GetInteractiveDefault 返回默认的交互模式
func GetInteractiveDefault() bool {
	// 从配置中获取值
	config, err := LoadConfig()
	if err != nil {
		// 出错时返回默认值
		return false
	}

	// 根据配置返回交互模式
	switch config.InteractiveMode {
	case "always":
		return true
	case "never":
		return false
	case "confirm":
		return true // confirm模式也启用交互
	default:
		return false
	}
}
