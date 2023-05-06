package models

type FightItem struct {
	FightToken  string  `bson:"fight_token"`
	Roles       []int64 `bson:"roles"`
	FightStatus int32   `bson:"fight_status"` // 0:未开始 1:已准备 2:战斗中 3:战斗结束
}
