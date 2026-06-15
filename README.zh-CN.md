# imgup

[English](README.md) | [简体中文](README.zh-CN.md)

[![CI](https://github.com/dengshu2/imgup/actions/workflows/ci.yml/badge.svg)](https://github.com/dengshu2/imgup/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/dengshu2/imgup.svg)](https://pkg.go.dev/github.com/dengshu2/imgup)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

一个轻量、零依赖的命令行工具,把图片上传到 **Cloudflare R2** 存储桶并输出公开 URL。它通过 `aws-sdk-go-v2` 直连 S3 API——不需要 Node、Python、GUI 或常驻后台进程。编译成单个静态二进制,丢进 `PATH` 即可用。

`imgup` 为脚本化而设计:**公开 URL 走 stdout,其余信息(进度、警告)走 stderr**,因此能干净地嵌入管道和上层薄封装。

```console
$ imgup screenshot.png
https://img.example.com/2026/06/9f86d081884c7d659a2feaa0c55ad015.png
```

## 特性

- 直连 R2 / S3 上传(`region=auto` + 自定义 endpoint),不依赖任何第三方上传器运行时。
- 内容寻址、**幂等**命名:对象 key 内嵌文件 md5,重复上传相同字节直接返回已有 URL。
- 公开 URL 恒为 `public_base + "/" + key`(自定义域名),不碰 `r2.dev`。
- 多 **profile**,便于管理多个桶。
- secret 解析链:**命令行 flag → 环境变量 → 配置文件 → macOS Keychain**。
- 输出格式:裸 URL、Markdown 或 HTML;可选 `--copy` 写入剪贴板。

## 安装

```sh
go install github.com/dengshu2/imgup@latest
```

或从源码构建:

```sh
git clone https://github.com/dengshu2/imgup
cd imgup
go build -o imgup .
```

## 快速开始

```sh
imgup init       # 交互式:生成 ~/.config/imgup/config.toml
imgup doctor     # 解析配置并真连一次 R2 验证
imgup photo.jpg  # 上传并输出 URL
```

## 配置

配置文件位于 `~/.config/imgup/config.toml`(可用 `--config` 或 `$IMGUP_CONFIG` 覆盖;遵循 `$XDG_CONFIG_HOME`),以 `0600` 权限写入。

```toml
default_profile = "default"

[profiles.default]
backend      = "r2"
endpoint     = "https://<account-id>.r2.cloudflarestorage.com"
bucket       = "img-disk"
public_base  = "https://img.example.com"
key_template = "{year}/{month}/{md5}.{ext}"
access_key_id = "<access-key-id>"

# secret 二选一:
use_keychain = true            # 从 OS keychain 读取(macOS 默认)
# secret_access_key = "..."    # 或内联存储(便携;请保持文件 0600)
```

### secret 存在哪

`imgup` 从不强制把 secret 写进配置文件。按如下顺序解析,命中即止:

| 优先级 | 来源 | 方式 |
|-------|------|------|
| 1 | flag | `--secret-access-key` |
| 2 | 环境变量 | `IMGUP_SECRET_ACCESS_KEY` |
| 3 | 配置文件 | 内联 `secret_access_key` |
| 4 | OS keychain | 当 `use_keychain = true`(service `imgup`,account = profile 名) |

**推荐:** 在 macOS 上让 `imgup init` 把 secret 存进 Keychain,配置文件就不含 secret,可安全同步/备份。Linux/CI 用环境变量或 `0600` 文件里的内联 secret(与 `~/.aws/credentials` 同模型)。若发现 group/world 可读的文件里存了内联 secret,`imgup doctor` 会告警。

### key 模板占位符

`{year}` `{month}` `{day}` `{md5}` `{ext}` `{filename}`。默认 `{year}/{month}/{md5}.{ext}` 生成 `2026/06/<md5>.png`。

### 环境变量

`IMGUP_CONFIG`、`--profile`(等价于选 profile)、`IMGUP_ENDPOINT`、`IMGUP_BUCKET`、`IMGUP_PUBLIC_BASE`、`IMGUP_KEY_TEMPLATE`、`IMGUP_ACCESS_KEY_ID`、`IMGUP_SECRET_ACCESS_KEY`、`IMGUP_BACKEND`。

## 用法

```sh
imgup <file>                 # 上传并把 URL 输出到 stdout
imgup -f md photo.png        # ![photo](https://.../photo.png)
imgup -f html photo.png      # <img src="..." alt="photo">
imgup --copy photo.png       # 顺便把 URL 复制到剪贴板
imgup --profile blog a.png   # 使用非默认 profile
imgup a.png b.jpg c.gif      # 多文件,每行一个 URL
cat shot.png | imgup -       # 从 stdin 读单张图(接截图管道)
```

flag:`-f, --format {url|md|html}` · `--copy` · `--profile` · `--config` · 设置覆盖项(`--endpoint`、`--bucket`、`--public-base`、`--key-template`、`--access-key-id`、`--secret-access-key`)。

### 命令

- `imgup init` —— 交互式创建/更新 profile。
- `imgup doctor` —— 解析当前 profile,显示每个值的来源,并真连一次 R2 验证。
- `imgup <file>...` —— 上传(默认动作)。`-` 从 stdin 读一张图。
- `imgup ls [prefix]` —— 列出对象;`-l` 显示大小 + 修改时间,`--url` 输出公开 URL,`--limit N` 限制数量。
- `imgup rm <key|url>...` —— 删除对象(接受 key 或完整公开 URL);`-y` 跳过确认。
- `imgup exists <file>` —— 检查某文件内容是否已存储;存在则输出 URL 并 exit 0,否则非 0 退出。
- `imgup skill install|print` —— 管理内置的 agent skill(见下)。

## 配套 skill

imgup 自带一个 **agent skill** —— 一份简短的说明文件,教 AI 编码 agent(Claude Code、Cursor……)何时、如何调用 imgup。它被 embed 进二进制,因此安装时**零网络、零外部文件**:

```sh
imgup skill install                    # 安装到所有检测到的 agent skill 目录
imgup skill install --to ~/.agents/skills   # 或指定目录(可重复)
imgup skill print                      # 把 SKILL.md 打到 stdout
```

源文件在 [`skill/SKILL.md`](skill/SKILL.md)。`install` 会写到 `<dir>/imgup/SKILL.md`;不带 `--to` 时,自动 fan-out 到所有存在的常见 agent skill 目录(`~/.agents/skills`、`~/.claude/skills`)。

## 路线图

存储层是接口(`storage.Uploader`),为未来非 R2 / S3 后端留出空间而无需改动命令层。可能的下一步:glob 通配、按 profile 的默认输出格式、以及基于 `imgup skill` 的薄 AI 封装。

## 许可证

[MIT](LICENSE)
