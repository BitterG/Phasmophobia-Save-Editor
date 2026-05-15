<p align="center">
  <img src="build/appicon.png" width="128" />
</p>
# Phasmophobia 存档修改器

基于 Wails v2 + Vue 3 构建的桌面端 Phasmophobia（恐鬼症）存档编辑器。支持读取、修改和写回游戏存档文件中的转生等级、玩家等级和金币数量。

## 功能特性

- **读取存档**——自动定位并解密 Phasmophobia 的加密存档文件（`SaveFile.txt`）
- **编辑三大属性**：
  - 转生等级（Prestige）：`0 ~ 3999`
  - 玩家等级（NewLevel）：`1 ~ 9999`
  - 金币（PlayersMoney）：`0 ~ 999,999,999`
- **保存修改**——修改后重新加密并写回存档，修改前自动备份明文数据
- **自定义深色主题 UI**——无边框窗口，数字滚动动画，扫光特效

## 技术原理

### 存档加密（EasySave3）

Phasmophobia 使用 Unity 插件 **EasySave3（ES3）** 保存存档数据。加密流程：

1. **密钥派生**：固定密钥 `t36gref9u84y7f43g` 经过 **PBKDF2-SHA1**（100 次迭代）派生 16 字节 AES 密钥
2. **加密算法**：**AES-CBC** 模式，PKCS7 填充
3. 加密文件存储于 `%USERPROFILE%\AppData\LocalLow\Kinetic Games\Phasmophobia\SaveFile.txt`

### 非标准 JSON 处理

游戏的 JSON 格式与 Go 标准库 `encoding/json` 不完全兼容，本项目采用**占位符替换法**处理：

- **`Items` 字段**：以对象作为 Map 的 key（标准 JSON 不允许），读取时提取原始内容并用占位符替换，保存时还原
- **`playedMaps.value` 字段**：包含复杂内联对象，同样使用占位符保留原始结构

### ES3 数值格式

存档中的数值字段使用 ES3 类型标记：

```json
{
  "Prestige": { "__type": "int", "value": 0 },
  "NewLevel": { "__type": "int", "value": 1 },
  "PlayersMoney": { "__type": "int", "value": 999999999 }
}
```

程序只修改 `value` 字段，保留 `__type` 标记不动。

## 技术栈

| 层级 | 技术 |
|------|------|
| 框架 | [Wails v2](https://wails.io/)（Go 桌面应用框架） |
| 后端 | Go 1.23 + `golang.org/x/crypto` |
| 前端 | Vue 3 + Vite |
| 加密 | AES-CBC + PBKDF2-SHA1 |

## 快速开始

### 环境要求

- Go 1.23+
- Node.js 16+
- Wails CLI（`go install github.com/wailsapp/wails/v2/cmd/wails@latest`）

### 开发模式

```bash
wails dev
```

此命令会同时启动 Vite 热重载开发服务器和 Go 后端，前端修改即时生效。

### 构建可执行文件

```bash
wails build
```

构建产物为单文件 `.exe`，无需任何运行时依赖即可运行。

## 项目结构

```
├── main.go                  # 应用入口，窗口配置，资源嵌入
├── app.go                   # 核心业务逻辑（读取/写入/校验存档）
├── es3_cipher.go            # ES3 加密/解密实现（PBKDF2 + AES-CBC）
├── es3_retool.go            # 非标准 JSON 修复工具（占位符替换/还原）
├── wails.json               # Wails 项目配置
├── go.mod / go.sum          # Go 依赖管理
├── build/                   # 构建资源（图标、清单、安装程序）
├── frontend/
│   ├── index.html           # HTML 入口
│   ├── package.json         # 前端依赖
│   ├── vite.config.js       # Vite 配置
│   └── src/
│       ├── main.js          # Vue 挂载入口
│       ├── App.vue          # 根组件
│       ├── style.css        # 全局样式（深色主题、字体）
│       ├── components/
│       │   ├── HelloWorld.vue   # 主界面（标题栏、存档卡片、按钮）
│       │   └── RollNumber.vue   # 数字滚动动画组件
│       └── assets/
│           ├── fonts/       # Nunito 字体
│           └── images/      # Logo 图片
```

## 核心代码说明

### `app.go`——业务逻辑

提供三个 Wails 绑定方法供前端调用：

- `LoadSave()`——读取加密存档 → 解密 → 修复 JSON → 解析 → 返回 `SaveData` 给前端
- `ApplySave(data)`——校验范围 → 修改字段 → 序列化 → 还原 JSON → 加密 → 写回文件
- `GetSavePath()`——返回存档文件路径

### `es3_cipher.go`——加密解密

对应原 Python 实现 `es3_decrypt` / `es3_encrypt`：

- `es3Decrypt(encData, key)`：PBKDF2 派生密钥 → AES-CBC 解密 → 去 PKCS7 填充
- `es3Encrypt(plainData, key)`：PKCS7 填充 → 生成随机 IV → AES-CBC 加密

### `es3_retool.go`——JSON 修复

- `extractAndReplaceItems()`：提取并替换 `Items` 非标准 JSON 片段
- `getOriginStr()` + `correctionJSON()`：提取并替换 `playedMaps.value` 内容
- `restoreItems()` + `recoverJSON2Str()`：写回时还原原始内容

## 注意事项

- 修改前会自动在存档同目录下生成 `SaveFile_backup.txt`（明文备份），如需恢复可手动处理
- 数值范围已做严格校验，超出范围将拒绝保存
- 仅支持 Windows 平台（存档路径使用 `USERPROFILE` 环境变量）
- 本工具仅供学习交流使用

## 作者

**苦瓜**（gourd_bitter）
