package main

import (
	"fmt"
	"regexp"
	"strings"
)

// placeholderValue 用于临时替换 playedMaps 的 value，避免其中的嵌套结构干扰 JSON 解析
// 对应 Python: Placeholder_value = '"placeholder_114514"'
const placeholderValue = `"placeholder_114514"`

// getOriginStr 提取 playedMaps.value 的原始内容（用于后续还原）
// 对应 Python: reTool.getOrginStr(str_json)
func getOriginStr(strJSON string) (string, error) {
	pattern := regexp.MustCompile(`(?s)"playedMaps"\s*:\s*\{[^}]+"value"\s*:\s*(\{[^}]+\})`)
	match := pattern.FindStringSubmatch(strJSON)
	if match == nil {
		return "", fmt.Errorf("未找到 'playedMaps' 的 'value' 字段")
	}
	return match[1], nil
}

// correctionJSON 将 playedMaps.value 替换为占位符，使整体 JSON 可被标准解析器解析
// 对应 Python: reTool.correction_json(str_json)
func correctionJSON(strJSON string) (string, error) {
	pattern := regexp.MustCompile(`(?s)("playedMaps"\s*:\s*\{[^}]+"value"\s*:\s*)\{[^}]+\}`)
	match := pattern.FindStringSubmatch(strJSON)
	if match == nil {
		return "", fmt.Errorf("未找到 'playedMaps' 部分")
	}
	prefix := match[1]
	result := pattern.ReplaceAllString(strJSON, prefix+placeholderValue)
	return result, nil
}

// recoverJSON2Str 将修改后的 JSON 对象序列化，并还原占位符为原始 playedMaps.value
// 对应 Python: recoverJson2Str(json_obj, orginStr)
func recoverJSON2Str(jsonStr string, originStr string) string {
	return strings.ReplaceAll(jsonStr, placeholderValue, originStr)
}
