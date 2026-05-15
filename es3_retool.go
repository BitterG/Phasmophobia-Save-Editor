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

// itemsPlaceholder 用于临时替换 Items 的 value，避免其中对象作为字典 key 的非标准 JSON 干扰解析
const itemsPlaceholder = `"placeholder_items_114514"`

// extractAndReplaceItems 找到 "Items" : {...} 区块，将其 value 替换为占位符
// Items 的 value 使用了 {obj}:{obj} 格式（对象作为 key），不是合法 JSON
func extractAndReplaceItems(strJSON string) (fixedJSON string, itemsOrigin string, err error) {
	re := regexp.MustCompile(`"Items"\s*:\s*`)
	loc := re.FindStringIndex(strJSON)
	if loc == nil {
		// 没有 Items 字段，无需处理
		return strJSON, "", nil
	}

	// 定位到第一个 {（Items dict 的起始括号）
	start := loc[1]
	for start < len(strJSON) && strJSON[start] != '{' {
		start++
	}
	if start >= len(strJSON) {
		return "", "", fmt.Errorf("未找到 Items 的起始括号")
	}

	// 括号计数找到匹配的 }
	count := 0
	end := start
	for end < len(strJSON) {
		if strJSON[end] == '{' {
			count++
		} else if strJSON[end] == '}' {
			count--
			if count == 0 {
				end++
				break
			}
		}
		end++
	}
	if count != 0 {
		return "", "", fmt.Errorf("Items 区块括号不匹配")
	}

	// itemsOrigin 只包含 value 部分（不含 "Items" : 前缀）
	itemsOrigin = strJSON[start:end]

	// 替换整个 "Items" : {...} 为 "Items" : "placeholder_items_114514"
	fixedJSON = strJSON[:loc[0]] + `"Items" : ` + itemsPlaceholder + strJSON[end:]

	return fixedJSON, itemsOrigin, nil
}

// restoreItems 将占位符还原为原始 Items value
func restoreItems(jsonStr string, itemsOrigin string) string {
	if itemsOrigin == "" {
		return jsonStr
	}
	return strings.ReplaceAll(jsonStr, itemsPlaceholder, itemsOrigin)
}
