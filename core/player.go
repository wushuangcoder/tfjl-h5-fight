package core

import (
	"tfjl-h5-fight/iface"
	"tfjl-h5-fight/models"
	"tfjl-h5-fight/protocols"
)

type Player struct {
	PID                      int64
	ShowID                   string
	Nickname                 string
	Conn                     iface.IConnection
	Key                      uint8
	Level                    int32
	BIndulge                 bool
	IndulgeTime              int32
	IndulgeDayOnlineTime     int32
	ForbidLoginTimeRemaining int32
	FightInfo                protocols.T_RoleAbstract
	FightToken               string
	FightPattern             int32
	FightType                int32
	FightStatus              int32 // 对战状态：1：正常，2：对战中，3：对战结束
}

func NewPlayer(role models.Role, conn iface.IConnection) (*Player, error) {
	p := &Player{
		PID:                      role.RoleID,
		ShowID:                   role.StrID,
		Nickname:                 role.RoleName,
		Conn:                     conn,
		Key:                      role.Key,
		Level:                    role.Level,
		BIndulge:                 role.BIndulge,
		IndulgeTime:              role.IndulgeTime,
		IndulgeDayOnlineTime:     role.IndulgeDayOnlineTime,
		ForbidLoginTimeRemaining: role.ForbidLoginTimeRemaining,
	}
	return p, nil
}

// 广播对战消息
func (p *Player) BroadCastFightMsg(msgType int, protocolNum uint32, data []byte) {
	otherRoleID := p.FightInfo.RoleID
	otherPlayer := WorldMgrObj.GetPlayerByPID(otherRoleID)
	if otherPlayer != nil {
		otherPlayer.Conn.SendMessage(msgType, protocolNum, data)
	}
}

// 玩家下线
func (p *Player) Offline() {
	WorldMgrObj.RemovePlayerByPID(p.PID)
}
