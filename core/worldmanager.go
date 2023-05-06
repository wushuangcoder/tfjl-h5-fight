package core

import (
	"bytes"
	"encoding/binary"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

/*
当前游戏世界总管理模块
*/
type WorldManager struct {
	// 当前全部在线Players集合
	Players map[int64]*Player
	Client  *websocket.Conn
	// 保护Players集合的锁
	pLock sync.RWMutex
}

// 提供一个对外的世界管理模块引用（全局）
var WorldMgrObj *WorldManager

func init() {
	WorldMgrObj = &WorldManager{
		Players: make(map[int64]*Player),
	}
}

// 添加一个玩家
func (wm *WorldManager) AddPlayer(player *Player) {
	wm.pLock.Lock()
	wm.Players[player.PID] = player
	wm.pLock.Unlock()
}

// 删除一个玩家
func (wm *WorldManager) RemovePlayerByPID(pid int64) {
	wm.pLock.Lock()
	delete(wm.Players, pid)
	wm.pLock.Unlock()
}

// 通过玩家ID查询Player对象
func (wm *WorldManager) GetPlayerByPID(pid int64) *Player {
	wm.pLock.RLock()
	defer wm.pLock.RUnlock()

	return wm.Players[pid]
}

// 获取全部在线玩家
func (wm *WorldManager) GetAllPlayers() []*Player {
	wm.pLock.RLock()
	defer wm.pLock.RUnlock()

	players := make([]*Player, 0)

	for _, player := range wm.Players {
		players = append(players, player)
	}

	return players
}

func (wm *WorldManager) SendMsg(protocolNum int32, data []byte) {
	buffer := new(bytes.Buffer)
	var length uint32 = uint32(8 + len(data))
	err := binary.Write(buffer, binary.LittleEndian, length)
	if err != nil {
		logrus.Error(err)
		return
	}
	err = binary.Write(buffer, binary.LittleEndian, uint32(protocolNum))
	if err != nil {
		logrus.Error(err)
		return
	}
	err = binary.Write(buffer, binary.LittleEndian, data)
	if err != nil {
		logrus.Error(err)
		return
	}
	if err = wm.Client.WriteMessage(websocket.BinaryMessage, buffer.Bytes()); err != nil {
		logrus.Error(err)
		return
	}
}
