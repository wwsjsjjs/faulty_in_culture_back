# ai


# 会话管理
列表查+会话增删查改
GET    /api/chat/sessions          # 会话列表（需实现）
POST   /api/chat/sessions          # 创建会话（已实现，改名）
GET    /api/chat/sessions/:id      # 会话详情（可选）
PUT    /api/chat/sessions/:id      # 更新会话详情（新增）
DELETE /api/chat/sessions/:id      # 删除会话（新增）

# 消息管理
增删查，不需要改。
GET    /api/chat/sessions/:id/messages       # 消息历史（已实现，改路径）
POST   /api/chat/sessions/:id/messages       # 发送消息（已实现，改路径）
DELETE /api/chat/messages/:id                 # 撤回消息（已实现，改路径）
