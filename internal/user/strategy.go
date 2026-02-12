package user

// ============================================================
// 排行榜策略模式 (Strategy Pattern)
// 设计模式：策略模式 - 定义算法族，分别封装，让它们可以互相替换
// 职责：封装不同排行榜类型的分数字段选择策略
// 优点：
// 1. 消除大量的if-else/switch-case语句
// 2. 符合开闭原则，新增排行榜类型只需添加新策略
// 3. 策略可以独立变化和测试
// ============================================================

// ScoreStrategy 分数策略接口
type ScoreStrategy interface {
	// GetScore 获取分数
	GetScore(user *Entity) int
	// SetScore 设置分数
	SetScore(user *Entity, score int)
	// GetFieldName 获取数据库字段名（用于ORDER BY）
	GetFieldName() string
}

// ============================================================
// 具体策略实现
// ============================================================

// score1Strategy 排行榜1策略
type score1Strategy struct{}

func (s *score1Strategy) GetScore(user *Entity) int        { return user.Score1 }
func (s *score1Strategy) SetScore(user *Entity, score int) { user.Score1 = score }
func (s *score1Strategy) GetFieldName() string             { return "score1" }

// score2Strategy 排行榜2策略
type score2Strategy struct{}

func (s *score2Strategy) GetScore(user *Entity) int        { return user.Score2 }
func (s *score2Strategy) SetScore(user *Entity, score int) { user.Score2 = score }
func (s *score2Strategy) GetFieldName() string             { return "score2" }

// score3Strategy 排行榜3策略
type score3Strategy struct{}

func (s *score3Strategy) GetScore(user *Entity) int        { return user.Score3 }
func (s *score3Strategy) SetScore(user *Entity, score int) { user.Score3 = score }
func (s *score3Strategy) GetFieldName() string             { return "score3" }

// score4Strategy 排行榜4策略
type score4Strategy struct{}

func (s *score4Strategy) GetScore(user *Entity) int        { return user.Score4 }
func (s *score4Strategy) SetScore(user *Entity, score int) { user.Score4 = score }
func (s *score4Strategy) GetFieldName() string             { return "score4" }

// score5Strategy 排行榜5策略
type score5Strategy struct{}

func (s *score5Strategy) GetScore(user *Entity) int        { return user.Score5 }
func (s *score5Strategy) SetScore(user *Entity, score int) { user.Score5 = score }
func (s *score5Strategy) GetFieldName() string             { return "score5" }

// score6Strategy 排行榜6策略
type score6Strategy struct{}

func (s *score6Strategy) GetScore(user *Entity) int        { return user.Score6 }
func (s *score6Strategy) SetScore(user *Entity, score int) { user.Score6 = score }
func (s *score6Strategy) GetFieldName() string             { return "score6" }

// score7Strategy 排行榜7策略
type score7Strategy struct{}

func (s *score7Strategy) GetScore(user *Entity) int        { return user.Score7 }
func (s *score7Strategy) SetScore(user *Entity, score int) { user.Score7 = score }
func (s *score7Strategy) GetFieldName() string             { return "score7" }

// score8Strategy 排行榜8策略
type score8Strategy struct{}

func (s *score8Strategy) GetScore(user *Entity) int        { return user.Score8 }
func (s *score8Strategy) SetScore(user *Entity, score int) { user.Score8 = score }
func (s *score8Strategy) GetFieldName() string             { return "score8" }

// score9Strategy 排行榜9策略
type score9Strategy struct{}

func (s *score9Strategy) GetScore(user *Entity) int        { return user.Score9 }
func (s *score9Strategy) SetScore(user *Entity, score int) { user.Score9 = score }
func (s *score9Strategy) GetFieldName() string             { return "score9" }

// ============================================================
// 策略工厂 (Factory Pattern)
// 设计模式：工厂模式 - 根据参数创建不同的策略实例
// ============================================================

// strategyFactory 策略工厂 (单例模式)
var strategyFactory = map[int]ScoreStrategy{
	1: &score1Strategy{},
	2: &score2Strategy{},
	3: &score3Strategy{},
	4: &score4Strategy{},
	5: &score5Strategy{},
	6: &score6Strategy{},
	7: &score7Strategy{},
	8: &score8Strategy{},
	9: &score9Strategy{},
}

// GetScoreStrategy 获取分数策略
// 设计模式：工厂方法 + 单例模式
// 参数：rankType 排行榜类型(1-9)
// 返回：对应的策略实例，如果类型无效返回nil
func GetScoreStrategy(rankType int) ScoreStrategy {
	return strategyFactory[rankType]
}

// ValidateRankType 验证排行榜类型是否有效
func ValidateRankType(rankType int) bool {
	return rankType >= 1 && rankType <= 9
}
