package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// KEY 对应 Python: KEY = b't36gref9u84y7f43g'
var saveKey = []byte("t36gref9u84y7f43g")

// SaveData 暴露给前端的存档数据结构
type SaveData struct {
	RebirthLevel int `json:"rebirthLevel"`
	Level        int `json:"level"`
	Gold         int `json:"gold"`
}

// App struct
type App struct {
	ctx context.Context

	// 内部状态，对应 Python 全局变量 orginStr / json_obj
	originStr string
	jsonObj   map[string]interface{}
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// localLowDir 返回当前用户的 AppData\LocalLow 目录。
// 使用 USERPROFILE 环境变量（Windows 自动设置，支持中文/特殊字符用户名），
// 比手动拼接 "C:\Users\<username>" 更可靠。
func localLowDir() string {
	// USERPROFILE 由 Windows 直接设置，值形如 C:\Users\张三，无需自行拼接
	profile := os.Getenv("USERPROFILE")
	if profile == "" {
		// 极少数情况下的兜底：HOMEDRIVE + HOMEPATH
		profile = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
	}
	return filepath.Join(profile, "AppData", "LocalLow")
}

// savePath 返回 Phasmophobia 存档文件路径
// 对应 Python: SaveCommonPath
func savePath() string {
	return filepath.Join(localLowDir(), "Kinetic Games", "Phasmophobia", "SaveFile.txt")
}

// backupPath 返回备份文件路径
func backupPath() string {
	return filepath.Join(localLowDir(), "Kinetic Games", "Phasmophobia", "SaveFile_backup.txt")
}

// LoadSave 读取并解密存档，返回 SaveData 给前端
// 对应 Python: getJson() + 读取 json_obj 中的三个字段
func (a *App) LoadSave() (SaveData, error) {
	path := savePath()

	// 读取加密二进制文件
	encData, err := os.ReadFile(path)
	if err != nil {
		return SaveData{}, fmt.Errorf("读取存档文件失败: %w", err)
	}

	// AES-CBC 解密
	decData, err := es3Decrypt(encData, saveKey)
	if err != nil {
		return SaveData{}, fmt.Errorf("解密失败: %w", err)
	}

	// 写备份
	_ = os.WriteFile(backupPath(), decData, 0644)

	strJSON := string(decData)

	// 提取 playedMaps.value 原始内容（保留用于还原）
	origin, err := getOriginStr(strJSON)
	if err != nil {
		return SaveData{}, fmt.Errorf("提取 playedMaps 失败: %w", err)
	}
	a.originStr = origin

	// 将 playedMaps.value 替换为占位符，使 JSON 可被标准解析
	fixedJSON, err := correctionJSON(strJSON)
	if err != nil {
		return SaveData{}, fmt.Errorf("修正 JSON 失败: %w", err)
	}

	// 解析 JSON
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(fixedJSON), &obj); err != nil {
		return SaveData{}, fmt.Errorf("解析 JSON 失败: %w", err)
	}
	a.jsonObj = obj

	// 读取目标字段
	data := SaveData{
		RebirthLevel: getIntField(obj, "Prestige"),
		Level:        getIntField(obj, "NewLevel"),
		Gold:         getIntField(obj, "PlayersMoney"),
	}

	return data, nil
}

// ApplySave 将前端传入的修改写回存档并加密保存
// 对应 Python: branch_1/2/3() + save()
func (a *App) ApplySave(data SaveData) error {
	if a.jsonObj == nil {
		return fmt.Errorf("请先读取存档")
	}

	// 校验范围（对应 Python 各 branch 的范围检查）
	if data.RebirthLevel < 0 || data.RebirthLevel > 20 {
		return fmt.Errorf("转生等级必须在 0-20 之间")
	}
	if data.Level < 1 || data.Level > 9999 {
		return fmt.Errorf("等级必须在 1-9999 之间")
	}
	if data.Gold < 0 || data.Gold > 9999999 {
		return fmt.Errorf("金币必须在 0-9999999 之间")
	}

	// 修改 json_obj 中对应字段的 value
	setIntField(a.jsonObj, "Prestige", data.RebirthLevel)
	setIntField(a.jsonObj, "NewLevel", data.Level)
	setIntField(a.jsonObj, "PlayersMoney", data.Gold)

	// 序列化回 JSON 字符串（带缩进，对应 Python json.dumps(indent=4)）
	jsonBytes, err := json.MarshalIndent(a.jsonObj, "", "    ")
	if err != nil {
		return fmt.Errorf("序列化 JSON 失败: %w", err)
	}

	// 还原 playedMaps.value 占位符为原始内容
	finalStr := recoverJSON2Str(string(jsonBytes), a.originStr)

	// AES-CBC 加密
	encData, err := es3Encrypt([]byte(finalStr), saveKey)
	if err != nil {
		return fmt.Errorf("加密失败: %w", err)
	}

	// 写回存档文件
	if err := os.WriteFile(savePath(), encData, 0644); err != nil {
		return fmt.Errorf("写入存档失败: %w", err)
	}

	return nil
}

// GetSavePath 返回存档路径，供前端展示
func (a *App) GetSavePath() string {
	return savePath()
}

// ── 内部工具函数 ──

// getIntField 从 ES3 格式的 map 中读取整数 value
// ES3 格式: { "FieldName": { "__type": "int", "value": 123 } }
func getIntField(obj map[string]interface{}, key string) int {
	entry, ok := obj[key]
	if !ok {
		return 0
	}
	entryMap, ok := entry.(map[string]interface{})
	if !ok {
		return 0
	}
	v, ok := entryMap["value"]
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	}
	return 0
}

// setIntField 向 ES3 格式的 map 中写入整数 value
func setIntField(obj map[string]interface{}, key string, value int) {
	entry, ok := obj[key]
	if !ok {
		return
	}
	entryMap, ok := entry.(map[string]interface{})
	if !ok {
		return
	}
	entryMap["value"] = float64(value)
}
