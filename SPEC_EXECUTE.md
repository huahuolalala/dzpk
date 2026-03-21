# [Spec] 德州扑克联机游戏 - 执行计划

**需求名称**：德州扑克联机游戏
**技术方案**：SPEC.md
**生成时间**：2026-03-21
**状态**：待确认

---

## TDD 开发规范（后端）

| 步骤 | 操作 |
|------|------|
| 1 | 先写测试（必须） |
| 2 | 运行测试 → FAIL（因为实现还没写） |
| 3 | 写实现代码 |
| 4 | 运行测试 → PASS |
| 5 | FAIL → 修复直到 PASS（禁止跳过） |

---

## 执行步骤

### Phase 1：基础框架搭建

**目标**：项目初始化，双向通信跑通

---

#### 步骤 1：初始化 Go 服务端项目

- **文件**：`server/go.mod`
- **依赖**：
  - `github.com/gin-gonic/gin v1.9.1`
  - `github.com/gorilla/websocket v1.5.1`
  - `github.com/google/uuid v1.5.0`
  - `github.com/go-playground/validator/v10 v10.16.0`
- **目录结构**：
  ```
  server/
  ├── cmd/server/main.go
  └── internal/
      ├── config/
      ├── handler/
      ├── service/
      ├── game/
      ├── model/
      └── repository/
  ```

---

#### 步骤 2：创建服务端入口

- **文件**：`server/cmd/server/main.go`
- **功能**：
  - 加载配置（WS_PORT、HTTP_PORT、LOG_LEVEL）
  - HTTP 服务（/health 健康检查）
  - WebSocket 监听启动

---

#### 步骤 3：实现 WebSocket Hub

- **文件**：`server/internal/handler/ws.go`
- **结构**：
  - `Hub`：管理所有连接、广播
  - `Client`：单个连接
  - `Room`：房间内的客户端映射
- **机制**：使用 sync.Map 或 sync.RWMutex

---

#### 步骤 4：定义消息模型

- **文件**：`server/internal/model/message.go`
- **消息类型**：
  - `create_room` / `join_room` / `leave_room` / `start_game`
  - `player_action`
  - `room_update` / `game_state` / `game_result` / `error`

---

#### 步骤 5：实现房间内存存储

- **文件**：`server/internal/repository/memory.go`
- **功能**：
  - `GenerateRoomCode()`：6位数字，唯一
  - `CreateRoom()` / `GetRoom()` / `DeleteRoom()`
  - `AddPlayer()` / `RemovePlayer()`：线程安全

---

#### 步骤 6：实现房间消息处理

- **文件**：`server/internal/handler/ws.go`
- **消息处理**：
  - `HandleCreateRoom`：生成房间码，返回给客户端
  - `HandleJoinRoom`：验证房间码，分配座位
  - `HandleLeaveRoom`：房主离开解散房间，否则只移除玩家
  - `HandleStartGame`：仅房主可发，验证至少2人

---

#### 步骤 7：初始化 Tauri 客户端项目

- **命令**：
  ```bash
  cd client
  npm create tauri-app@latest . -- --template vue-ts --manager npm --yes
  ```
- **目录结构**：
  ```
  client/
  ├── src/          # Rust
  ├── web/          # Vue 前端
  ├── Cargo.toml
  └── package.json
  ```

---

#### 步骤 8：配置 Tauri Rust WebSocket

- **文件**：`client/src/main.rs` + `client/src/websocket.rs`
- **功能**：WebSocket 客户端连接、发送、接收

---

#### 步骤 9：创建 Vue 前端基础结构

- **文件**：`client/web/src/`
- **核心文件**：
  - `App.vue`：路由视图切换
  - `views/HomeView.vue`：创建/加入房间入口
  - `stores/room.ts`：Pinia 房间状态
  - `composables/useWebSocket.ts`：WS 连接管理
  - `styles/main.css`：全局样式

---

#### 步骤 10：打通双向通信

- **验证点**：
  1. `curl http://localhost:8081/health` → `{"status":"ok"}`
  2. 服务端启动
  3. 客户端启动
  4. 客户端发送消息，服务端广播

---

### Phase 2：房间系统

**目标**：创建/加入/解散房间完整流程

---

#### 步骤 11：完善广播机制

- **文件**：`server/internal/handler/ws.go`
- **功能**：
  - `BroadcastRoomUpdate`：房间状态变更广播
  - `BroadcastRoomDismissed`：房间解散广播

---

#### 步骤 12：Vue 房间等待页面

- **文件**：`client/web/src/views/RoomView.vue`
- **功能**：
  - 显示房间码
  - 显示玩家列表（座位、名称、筹码）
  - 房主显示"开始游戏"按钮
  - 非房主显示等待状态

---

### Phase 3：核心牌局逻辑（TDD）

**目标**：完整一局德州扑克
**顺序**：先写测试 → 再写实现 → 直到测试通过

---

#### 步骤 13：扑克牌模型 + 测试

- **文件**：`server/internal/game/poker.go` + `_test.go`
- **测试用例**：`TestCard_NewCard`、`TestCard_String`、`TestRank_String`、`TestSuit_String`

---

#### 步骤 14：发牌逻辑 + 测试

- **文件**：`server/internal/game/deck.go` + `_test.go`
- **测试用例**：`TestNewDeck`、`TestDeck_Shuffle`、`TestDeck_Deal`、`TestDeck_DealAll`

---

#### 步骤 15：牌型判断 + 测试

- **文件**：`server/internal/game/judge.go` + `_test.go`
- **测试用例**：
  - `TestHandRank_RoyalFlush`（9）
  - `TestHandRank_StraightFlush`（8）
  - `TestHandRank_FourOfAKind`（7）
  - `TestHandRank_FullHouse`（6）
  - `TestHandRank_Flush`（5）
  - `TestHandRank_Straight`（4）包括 A-2-3-4-5
  - `TestHandRank_ThreeOfAKind`（3）
  - `TestHandRank_TwoPair`（2）
  - `TestHandRank_OnePair`（1）
  - `TestHandRank_HighCard`（0）
  - `TestCompareHands`

---

#### 步骤 16：游戏状态机 + 测试

- **文件**：`server/internal/game/action.go` + `_test.go`
- **结构**：`Phase`（preflop/flop/turn/river/showdown）、`Player`、`GameState`
- **测试用例**：`TestGameState_NewGame`、`TestGameState_ProcessAction_Check/Call/Raise/Fold/AllIn`、`TestGameState_NextPhase`

---

#### 步骤 17：结算逻辑 + 测试

- **文件**：`server/internal/game/settle.go` + `_test.go`
- **功能**：`DetermineWinners`、`Settle`
- **测试用例**：`TestSettle_TwoPlayers`、`TestSettle_SplitPot`、`TestSettle_AllInSidePot`

---

#### 步骤 18：房间存储测试

- **文件**：`server/internal/repository/memory_test.go`
- **测试用例**：`TestRoomStore_Create`、`TestRoomStore_Get`、`TestRoomStore_AddPlayer`、`TestRoomStore_RemovePlayer`、`TestRoomStore_Delete`、`TestRoomStore_MaxSeats`

---

#### 步骤 19：集成游戏逻辑到 Handler

- **文件**：`server/internal/handler/ws.go`
- **功能**：
  - `HandlePlayerAction`：处理玩家动作
  - `BroadcastGameState`：广播游戏状态
  - `BroadcastGameResult`：广播结算结果
  - 所有人行动完毕自动进入下一阶段

---

### Phase 4：UI 界面

**目标**：完整可玩的界面

---

#### 步骤 20：游戏主视图

- **文件**：`client/web/src/views/GameView.vue`
- **布局**：顶部信息栏、牌桌、玩家座位环、手牌区域、动作栏

---

#### 步骤 21：玩家卡片组件

- **文件**：`client/web/src/components/PlayerCard.vue`
- **显示**：名称、筹码、手牌（暗/明）、状态（弃牌/All-in）、当前回合高亮

---

#### 步骤 22：公共牌组件

- **文件**：`client/web/src/components/CommunityCards.vue`
- **显示**：5张牌位置（Flop x3, Turn x1, River x1）

---

#### 步骤 23：手牌组件

- **文件**：`client/web/src/components/HandCard.vue`
- **显示**：带花色、选中状态

---

#### 步骤 24：动作栏组件

- **文件**：`client/web/src/components/ActionBar.vue`
- **按钮**：过牌、跟注（显示金额）、加注（弹出输入）、弃牌、All-in
- **状态**：根据当前状态启用/禁用

---

#### 步骤 25：快捷键

- **文件**：`client/web/src/composables/useKeyboard.ts`
- **映射**：空格→过牌/跟注、A→加注、F→弃牌、1-9→选玩家、Esc→关闭弹窗

---

### Phase 5：测试与发布

**目标**：稳定可发布

---

#### 步骤 26：运行所有单元测试

```bash
cd server
go test ./... -v -count=1
go test ./... -cover
```

**验证**：全部通过

---

#### 步骤 27：配置 Tauri 打包

- **文件**：`client/src-tauri/tauri.conf.json`
- **配置**：应用名"德州扑克"、窗口大小、打包目标

---

#### 步骤 28：执行打包

```bash
cd client
npm run tauri build
```

**验证**：.exe 存在、包体积 < 50MB、可运行

---

## 风险汇总

| 风险点 | 级别 | 处理方式 |
|--------|------|----------|
| 牌型判断逻辑复杂 | 高 | TDD 测试覆盖所有牌型 |
| 状态转换可能遗漏边界 | 高 | 测试覆盖所有转换 |
| 多人 All-in 边池计算 | 高 | 充分测试各种边池场景 |
| Tauri 首次打包配置 | 中 | 参考官方文档 |

---

## 确认签字

- [x] 技术方案已确认（SPEC.md）
- [x] 风险点已确认
- [ ] 每步骤按序执行

---

*如无异议，请回复「确认」，我将开始按步骤执行*
