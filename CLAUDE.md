# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 提供代码库操作指南。

## 项目概述

德州扑克联机游戏 - 使用 Go 后端 + Tauri 桌面客户端实现的实时多人扑克游戏。

## 技术架构

```
客户端 (Tauri v2) ←→ WebSocket ←→ 服务端 (Go + gorilla/websocket)
  Rust + Vue 3                    Go + gin
  Pinia 状态管理                   内存房间存储
```

## 常用命令

### 服务端 (Go)
```bash
cd server
go build -o main ./cmd/server          # 编译
go test ./... -v                       # 运行所有测试
go test ./internal/game -v             # 运行指定包测试
go test -run TestCard_NewCard -v       # 运行单个测试
gofmt -l .                             # 检查代码格式（提交前必查）
```

### 客户端 (Tauri)
```bash
cd client
npm install                            # 安装依赖
npm run dev                            # 开发服务器 (Vue 前端)
npm run tauri dev                      # 完整 Tauri 开发模式
npm run tauri build                    # 生产构建
```

## TDD 开发流程（后端必须遵守）

1. 先编写测试文件 (`_test.go`)
2. 运行测试 → 预期 FAIL（因为实现还没写）
3. 编写实现代码
4. 运行测试 → 预期 PASS
5. 如果 FAIL → 持续修复直到 PASS（禁止跳过）

测试文件与实现文件在同一包目录下。

## 服务端目录结构

```
server/
├── cmd/server/main.go        # 入口：Hub 初始化、HTTP/WS 服务启动
├── internal/
│   ├── config/               # 环境变量配置 (HTTP_PORT, WS_PORT, LOG_LEVEL)
│   ├── game/                 # 核心扑克逻辑（必须 TDD）
│   │   ├── poker.go         # 扑克牌模型、花色、点数
│   │   ├── deck.go          # 洗牌、发牌
│   │   ├── judge.go         # 牌型判断 (皇家同花顺 → 高牌)
│   │   ├── action.go         # 游戏状态机、玩家动作
│   │   └── settle.go         # 底池分配、赢家结算
│   ├── handler/
│   │   └── ws.go            # WebSocket Hub、Client、房间消息处理
│   ├── model/                # Room、Player、Message 数据结构
│   └── repository/
│       └── memory.go         # 内存房间存储（线程安全）
```

## 游戏逻辑要点

- **阶段**: preflop → flop → turn → river → showdown
- **动作**: 过牌(check)、跟注(call)、加注(raise)、弃牌(fold)、全下(allin)
- **牌型**: 10 个等级（皇家同花顺 = 9，高牌 = 0）
- **结算**: 主底池 + 边池（处理多人全下场景）

## WebSocket 通信协议

消息格式为 JSON: `{"type": "action_type", "data": {...}}`

客户端→服务端: `join_room`、`leave_room`、`start_game`、`player_action`
服务端→客户端: `room_update`、`game_state`、`game_result`、`error`

## 客户端目录结构

```
client/
├── src/                      # Rust 后端 (WebSocket 客户端、Tauri 命令)
├── web/src/                  # Vue 3 前端
│   ├── views/               # HomeView、RoomView、GameView
│   ├── components/          # PokerTable、PlayerCard、HandCard、ActionBar
│   ├── stores/              # Pinia 状态管理 (room.ts、game.ts)
│   └── composables/         # useWebSocket、useKeyboard
```

## 环境变量

服务端:
- `HTTP_PORT` (默认: 8081)
- `WS_PORT` (默认: 8080)
- `LOG_LEVEL` (默认: info)
