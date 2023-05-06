package apis

import (
	"bytes"
	"math/rand"
	"tfjl-h5-fight/core"
	"tfjl-h5-fight/db"
	"tfjl-h5-fight/iface"
	"tfjl-h5-fight/models"
	"tfjl-h5-fight/net"
	"tfjl-h5-fight/protocols"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type FightRoleLoginRouter struct {
	net.BaseRouter
}

func (p *FightRoleLoginRouter) Handle(request iface.IRequest) {
	logrus.Info("*********************************对战登录************************************")

	var cFightRoleLogin = protocols.C_Fight_Role_Login{}
	cFightRoleLogin.Decode(bytes.NewBuffer(request.GetData()))
	// logrus.Infof("%#v", cFightRoleLogin)

	role := db.DbManager.FindRoleByRoleID(cFightRoleLogin.RoleID)
	if role == (models.Role{}) {
		logrus.Error("role not found")
		return
	}
	player, err := core.NewPlayer(role, request.GetConnection())
	if err != nil {
		logrus.Error("NewPlayer error:", err)
		return
	}
	for _, v := range cFightRoleLogin.MatchRoles {
		if v.RoleID != role.RoleID {
			player.FightInfo = v
			break
		}
	}
	player.FightToken = cFightRoleLogin.FightToken
	player.FightPattern = cFightRoleLogin.FightPattern
	player.FightType = cFightRoleLogin.FightType
	core.WorldMgrObj.AddPlayer(player)
	request.GetConnection().SetProperty("roleID", role.RoleID)
	request.GetConnection().SetProperty("key", role.Key)

	var sFightRoleLogin = protocols.S_Fight_Role_Login{
		Errorcode: 0,
		Key:       role.Key,
	}
	request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_Role_Login, sFightRoleLogin.Encode())
}

type FightLoadingReadyRouter struct {
	net.BaseRouter
}

func (p *FightLoadingReadyRouter) Handle(request iface.IRequest) {
	logrus.Info("**************************************对战加载准备*******************************************")
	roleID, err := request.GetConnection().GetProperty("roleID")
	if err != nil {
		logrus.Error("GetProperty error:", err)
		return
	}
	player := core.WorldMgrObj.GetPlayerByPID(roleID.(int64))

	var cFightLoadingReady = protocols.C_Fight_Loading_Ready{}
	cFightLoadingReady.Decode(bytes.NewBuffer(request.GetData()), player.Key)
	// logrus.Infof("%#v", cFightLoadingReady)

	fightItem := db.DbManager.FindFightItemByFightToken(player.FightToken)
	for _, v := range fightItem.Roles {
		if v == player.PID {
			if fightItem.FightStatus == 0 {
				player.FightStatus = 1
				db.DbManager.UpdateFightItem(bson.M{"fight_token": fightItem.FightToken}, bson.M{"$set": bson.M{"fight_status": 1}})
			} else if fightItem.FightStatus == 1 {
				player.FightStatus = 2
				opponentPlayer := core.WorldMgrObj.GetPlayerByPID(player.FightInfo.RoleID)
				opponentPlayer.FightStatus = 2
				db.DbManager.UpdateFightItem(bson.M{"fight_token": fightItem.FightToken}, bson.M{"$set": bson.M{"fight_status": 2}})
				var sFightLoadingReady = protocols.S_Fight_Loading_Ready{
					Errorcode: 0,
				}
				request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_Loading_Ready, sFightLoadingReady.Encode())
				player.BroadCastFightMsg(request.GetMsgType(), protocols.P_Fight_Loading_Ready, sFightLoadingReady.Encode())
				// 发送战斗开始消息
				var sFightFightStart = protocols.S_Fight_FightStart{}
				request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_FightStart, sFightFightStart.Encode())
				player.BroadCastFightMsg(request.GetMsgType(), protocols.P_Fight_FightStart, sFightFightStart.Encode())
			}
			break
		}
	}
}

type FightSilverSYNCRouter struct {
	net.BaseRouter
}

func (p *FightSilverSYNCRouter) Handle(request iface.IRequest) {
	logrus.Info("*************************************银币同步***********************************************")
	roleID, err := request.GetConnection().GetProperty("roleID")
	if err != nil {
		logrus.Error("GetProperty error:", err)
		return
	}
	player := core.WorldMgrObj.GetPlayerByPID(roleID.(int64))

	var cFightSilverSYNC = protocols.C_Fight_Silver_SYNC{}
	cFightSilverSYNC.Decode(bytes.NewBuffer(request.GetData()), player.Key)
	// logrus.Infof("%#v", cFightSilverSYNC)

	fightItem := db.DbManager.FindFightItemByFightToken(player.FightToken)

	if fightItem.FightStatus == 2 {
		// 进入战斗，同步银币数据
		var sFightSilverSYNC = protocols.S_Fight_Silver_SYNC{}
		sFightSilverSYNC.RoleID = cFightSilverSYNC.RoleID
		sFightSilverSYNC.Silver = cFightSilverSYNC.Silver
		sFightSilverSYNC.TimeFrame = cFightSilverSYNC.TimeFrame

		request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_Silver_SYNC, sFightSilverSYNC.Encode())
		player.BroadCastFightMsg(request.GetMsgType(), protocols.P_Fight_Silver_SYNC, sFightSilverSYNC.Encode())
	}
}

type FightRefreshCardCountSYNCRouter struct {
	net.BaseRouter
}

func (p *FightRefreshCardCountSYNCRouter) Handle(request iface.IRequest) {
	logrus.Info("*************************************刷新卡牌次数同步********************************************")
	roleID, err := request.GetConnection().GetProperty("roleID")
	if err != nil {
		logrus.Error("GetProperty error:", err)
		return
	}
	player := core.WorldMgrObj.GetPlayerByPID(roleID.(int64))

	var cFightRefreshCardCountSYNC = protocols.C_Fight_Refresh_Card_Count_SYNC{}
	cFightRefreshCardCountSYNC.Decode(bytes.NewBuffer(request.GetData()), player.Key)
	// logrus.Infof("%#v", cFightRefreshCardCountSYNC)

	fightItem := db.DbManager.FindFightItemByFightToken(player.FightToken)

	if fightItem.FightStatus == 2 {
		var sFightRefreshCardCountSYNC = protocols.S_Fight_Refresh_Card_Count_SYNC{}
		sFightRefreshCardCountSYNC.RoleID = cFightRefreshCardCountSYNC.RoleID
		sFightRefreshCardCountSYNC.RefreshNum = cFightRefreshCardCountSYNC.RefreshNum
		sFightRefreshCardCountSYNC.CardIdVec = cFightRefreshCardCountSYNC.CardIdVec

		request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_Refresh_Card_Count_SYNC, sFightRefreshCardCountSYNC.Encode())
		player.BroadCastFightMsg(request.GetMsgType(), protocols.P_Fight_Refresh_Card_Count_SYNC, sFightRefreshCardCountSYNC.Encode())
	}
}

type FightHeroAttrSYNCRouter struct {
	net.BaseRouter
}

func (p *FightHeroAttrSYNCRouter) Handle(request iface.IRequest) {
	logrus.Info("*********************************对战英雄属性同步***************************************")
	roleID, err := request.GetConnection().GetProperty("roleID")
	if err != nil {
		logrus.Error("GetProperty error:", err)
		return
	}
	player := core.WorldMgrObj.GetPlayerByPID(roleID.(int64))

	var cFightHeroAttrSYNC = protocols.C_Fight_Hero_Attr_SYNC{}
	cFightHeroAttrSYNC.Decode(bytes.NewBuffer(request.GetData()), player.Key)
	// logrus.Infof("%#v", cFightHeroAttrSYNC)

	fightItem := db.DbManager.FindFightItemByFightToken(player.FightToken)

	if fightItem.FightStatus == 2 {
		var sFightHeroAttrSYNC = protocols.S_Fight_Hero_Attr_SYNC{}
		sFightHeroAttrSYNC.RoleID = cFightHeroAttrSYNC.RoleID
		sFightHeroAttrSYNC.HeroAttrVec = cFightHeroAttrSYNC.HeroAttrVec

		// logrus.Info("发送对战英雄属性同步: ", sFightHeroAttrSYNC)
		request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_Hero_Attr_SYNC, sFightHeroAttrSYNC.Encode())
		player.BroadCastFightMsg(request.GetMsgType(), protocols.P_Fight_Hero_Attr_SYNC, sFightHeroAttrSYNC.Encode())
	}
}

type FightMonsterBloodSYNCRouter struct {
	net.BaseRouter
}

func (p *FightMonsterBloodSYNCRouter) Handle(request iface.IRequest) {
	logrus.Info("*********************************对战怪物血量同步*********************************")
	roleID, err := request.GetConnection().GetProperty("roleID")
	if err != nil {
		logrus.Error("GetProperty error:", err)
		return
	}
	player := core.WorldMgrObj.GetPlayerByPID(roleID.(int64))

	var cFightMonsterBloodSYNC = protocols.C_Fight_Monster_Blood_SYNC{}
	cFightMonsterBloodSYNC.Decode(bytes.NewBuffer(request.GetData()), player.Key)
	// logrus.Infof("%#v", cFightMonsterBloodSYNC)

	fightItem := db.DbManager.FindFightItemByFightToken(player.FightToken)

	if fightItem.FightStatus == 2 {

		var masterFightMonsterBloodSYNC = protocols.S_Fight_Monster_Blood_SYNC{
			TimeFrame:    cFightMonsterBloodSYNC.TimeFrame,
			SyncRoleID:   player.PID,
			MonsterBlood: make(map[int64]protocols.T_Monster_Blood_Struct),
		}
		var slaveFightMonsterBloodSYNC = protocols.S_Fight_Monster_Blood_SYNC{
			TimeFrame:    cFightMonsterBloodSYNC.TimeFrame,
			SyncRoleID:   player.PID,
			MonsterBlood: make(map[int64]protocols.T_Monster_Blood_Struct),
		}
		// bytesArr, _ := json.Marshal(cFightMonsterBloodSYNC)
		// logrus.Info("roleID: ", player.PID)
		// logrus.Info(string(bytesArr))
		for k, v := range cFightMonsterBloodSYNC.PlayerMonsterBlood {
			if k == 2 {
				// 车辆扣血
			} else if k == fightItem.Roles[1] {
				// 主机
				for k2, v2 := range v.MonsterBlood {
					if len(v2.MonsterBlood) == 0 {
						continue
					}
					if k2 == 2 {
						// 主机怪物扣血
						for k3, v3 := range v2.MonsterBlood {
							curBlood := v3.CurBlood - v3.DropBlood
							if curBlood > v3.MaxBlood {
								curBlood = v3.MaxBlood
							}
							var blood = protocols.T_Blood_Struct{
								DropBlood:      0,
								CurBlood:       curBlood,
								MaxBlood:       v3.MaxBlood,
								AddBlood:       0,
								Type:           v3.Type,
								HeroId:         v3.HeroId,
								ExtraDropMoney: v3.ExtraDropMoney,
								DamageType:     0,
								MaxDropBlood:   0,
							}
							v2.MonsterBlood[k3] = blood
							masterFightMonsterBloodSYNC.MonsterBlood[2] = v2
							slaveFightMonsterBloodSYNC.MonsterBlood[2] = v2
						}
					} else if k2 == fightItem.Roles[0] {
						if v3, ok := v2.MonsterBlood[2]; ok {
							if v3.DropBlood == v3.MaxBlood && v3.CurBlood == 0 && v3.DamageType == 0 {
								// 主机结束
								player.FightStatus = 3
								opponentPlayer := core.WorldMgrObj.GetPlayerByPID(player.FightInfo.RoleID)
								if opponentPlayer.FightStatus == 3 {
									fightItem.FightStatus = 3
									db.DbManager.UpdateFightItem(bson.M{"fight_token": fightItem.FightToken}, bson.M{"$set": bson.M{"fight_status": 3}})
									var sFightFightEnd = protocols.S_Fight_FightEnd{Reason: 0, WinRoleID: player.PID}
									request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_FightEnd, sFightFightEnd.Encode())
									opponentPlayer.Conn.SendMessage(request.GetMsgType(), protocols.P_Fight_FightEnd, sFightFightEnd.Encode())
								}
							} else {
								player.FightStatus = 2
							}
						}
					}
				}
			} else if k == fightItem.Roles[0] {
				// 从机
				for k2, v2 := range v.MonsterBlood {
					if len(v2.MonsterBlood) == 0 {
						continue
					}
					if k2 == 2 {
						// 从机怪物扣血
						for k3, v3 := range v2.MonsterBlood {
							curBlood := v3.CurBlood - v3.DropBlood
							if curBlood > v3.MaxBlood {
								curBlood = v3.MaxBlood
							}
							var blood = protocols.T_Blood_Struct{
								DropBlood:      0,
								CurBlood:       curBlood,
								MaxBlood:       v3.MaxBlood,
								AddBlood:       0,
								Type:           v3.Type,
								HeroId:         v3.HeroId,
								ExtraDropMoney: v3.ExtraDropMoney,
								DamageType:     0,
								MaxDropBlood:   0,
							}
							v2.MonsterBlood[k3] = blood
							masterFightMonsterBloodSYNC.MonsterBlood[2] = v2
							slaveFightMonsterBloodSYNC.MonsterBlood[2] = v2
						}
					} else if k2 == fightItem.Roles[0] {
						if v3, ok := v2.MonsterBlood[2]; ok {
							if v3.DropBlood == v3.MaxBlood && v3.CurBlood == 0 && v3.DamageType == 0 {
								// 从机结束
								player.FightStatus = 3
								opponentPlayer := core.WorldMgrObj.GetPlayerByPID(player.FightInfo.RoleID)
								if opponentPlayer.FightStatus == 3 {
									fightItem.FightStatus = 3
									db.DbManager.UpdateFightItem(bson.M{"fight_token": fightItem.FightToken}, bson.M{"$set": bson.M{"fight_status": 3}})
									var sFightFightEnd = protocols.S_Fight_FightEnd{Reason: 0, WinRoleID: player.PID}
									request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_FightEnd, sFightFightEnd.Encode())
									opponentPlayer.Conn.SendMessage(request.GetMsgType(), protocols.P_Fight_FightEnd, sFightFightEnd.Encode())
								}
							} else {
								player.FightStatus = 2
							}
						}
					}
				}
			}
		}
		// bytesArr, _ = json.Marshal(masterFightMonsterBloodSYNC)
		// logrus.Info("发送数据主机: ", string(bytesArr))
		// bytesArr, _ = json.Marshal(slaveFightMonsterBloodSYNC)
		// logrus.Info("发送数据从机: ", string(bytesArr))
		request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_Monster_Blood_SYNC, masterFightMonsterBloodSYNC.Encode())
		player.BroadCastFightMsg(request.GetMsgType(), protocols.P_Fight_Monster_Blood_SYNC, slaveFightMonsterBloodSYNC.Encode())
	}
}

type FightUpdateCarSYNCRouter struct {
	net.BaseRouter
}

func (p *FightUpdateCarSYNCRouter) Handle(request iface.IRequest) {
	logrus.Info("******************************对战更新战车同步***********************************")
	roleID, err := request.GetConnection().GetProperty("roleID")
	if err != nil {
		logrus.Error("GetProperty error:", err)
		return
	}
	player := core.WorldMgrObj.GetPlayerByPID(roleID.(int64))

	var cFightUpdateCarSYNC = protocols.C_Fight_Update_Car_SYNC{}
	cFightUpdateCarSYNC.Decode(bytes.NewBuffer(request.GetData()), player.Key)
	logrus.Infof("%#v", cFightUpdateCarSYNC)

	fightItem := db.DbManager.FindFightItemByFightToken(player.FightToken)

	if fightItem.FightStatus == 2 {
		var sFightUpdateCarSYNC = protocols.S_Fight_Update_Car_SYNC{}
		sFightUpdateCarSYNC.RoleID = cFightUpdateCarSYNC.RoleID
		sFightUpdateCarSYNC.UpdateCar = cFightUpdateCarSYNC.UpdateCar
		request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_Update_Car_SYNC, sFightUpdateCarSYNC.Encode())
		player.BroadCastFightMsg(request.GetMsgType(), protocols.P_Fight_Update_Car_SYNC, sFightUpdateCarSYNC.Encode())
	}
}

type FightUpdateHeroSYNCRouter struct {
	net.BaseRouter
}

func (p *FightUpdateHeroSYNCRouter) Handle(request iface.IRequest) {
	logrus.Info("***************************对战更新英雄同步********************************")
	roleID, err := request.GetConnection().GetProperty("roleID")
	if err != nil {
		logrus.Error("GetProperty error:", err)
		return
	}
	player := core.WorldMgrObj.GetPlayerByPID(roleID.(int64))

	var cFightUpdateHeroSYNC = protocols.C_Fight_Update_Hero_SYNC{}
	cFightUpdateHeroSYNC.Decode(bytes.NewBuffer(request.GetData()), player.Key)
	logrus.Infof("%#v", cFightUpdateHeroSYNC)

	fightItem := db.DbManager.FindFightItemByFightToken(player.FightToken)

	if fightItem.FightStatus == 2 {
		var sFightUpdateHeroSYNC = protocols.S_Fight_Update_Hero_SYNC{}
		sFightUpdateHeroSYNC.RoleID = cFightUpdateHeroSYNC.RoleID
		sFightUpdateHeroSYNC.UpdateHero = cFightUpdateHeroSYNC.UpdateHero

		request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_Update_Hero_SYNC, sFightUpdateHeroSYNC.Encode())
		player.BroadCastFightMsg(request.GetMsgType(), protocols.P_Fight_Update_Hero_SYNC, sFightUpdateHeroSYNC.Encode())
	}
}

type FightSellHeroSYNCRouter struct {
	net.BaseRouter
}

func (p *FightSellHeroSYNCRouter) Handle(request iface.IRequest) {
	logrus.Info("***************************对战出售英雄同步********************************")
	roleID, err := request.GetConnection().GetProperty("roleID")
	if err != nil {
		logrus.Error("GetProperty error:", err)
		return
	}
	player := core.WorldMgrObj.GetPlayerByPID(roleID.(int64))

	var cFightSellHeroSYNC = protocols.C_Fight_Sell_Hero_SYNC{}
	cFightSellHeroSYNC.Decode(bytes.NewBuffer(request.GetData()), player.Key)
	logrus.Infof("%#v", cFightSellHeroSYNC)

	fightItem := db.DbManager.FindFightItemByFightToken(player.FightToken)

	if fightItem.FightStatus == 2 {
		var sFightSellHeroSYNC = protocols.S_Fight_Sell_Hero_SYNC{}
		sFightSellHeroSYNC.RoleID = cFightSellHeroSYNC.RoleID
		sFightSellHeroSYNC.SellHero = cFightSellHeroSYNC.SellHero

		request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_Sell_Hero_SYNC, sFightSellHeroSYNC.Encode())
		player.BroadCastFightMsg(request.GetMsgType(), protocols.P_Fight_Sell_Hero_SYNC, sFightSellHeroSYNC.Encode())
	}
}

type FightSkillSYNCRouter struct {
	net.BaseRouter
}

func (p *FightSkillSYNCRouter) Handle(request iface.IRequest) {
	logrus.Info("*********************************对战技能同步********************************")
}

type FightOperateEquipSYNCRouter struct {
	net.BaseRouter
}

func (p *FightOperateEquipSYNCRouter) Handle(request iface.IRequest) {
	logrus.Info("*********************************操作装备同步********************************")
	roleID, err := request.GetConnection().GetProperty("roleID")
	if err != nil {
		logrus.Error("GetProperty error:", err)
		return
	}
	player := core.WorldMgrObj.GetPlayerByPID(roleID.(int64))

	var cFightOperateEquipSYNC = protocols.C_Fight_Operate_Equip_SYNC{}
	cFightOperateEquipSYNC.Decode(bytes.NewBuffer(request.GetData()), player.Key)
	logrus.Infof("%#v", cFightOperateEquipSYNC)

	fightItem := db.DbManager.FindFightItemByFightToken(player.FightToken)

	if fightItem.FightStatus == 2 {
		var sFightOperateEquipSYNC = protocols.S_Fight_Operate_Equip_SYNC{}
		sFightOperateEquipSYNC.Errorcode = 0
		sFightOperateEquipSYNC.OperateType = cFightOperateEquipSYNC.OperateType
		sFightOperateEquipSYNC.CarId = cFightOperateEquipSYNC.CarId
		sFightOperateEquipSYNC.Target = cFightOperateEquipSYNC.Target
		rand.Seed(time.Now().UnixNano())
		sFightOperateEquipSYNC.EquipItems = map[int64]protocols.EquipItem{
			cFightOperateEquipSYNC.RoleID: {
				ItemID:  int32(rand.Intn(4) + 501),
				ItemNum: 1,
			},
		}

		request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_Operate_Equip_SYNC, sFightOperateEquipSYNC.Encode())
		player.BroadCastFightMsg(request.GetMsgType(), protocols.P_Fight_Operate_Equip_SYNC, sFightOperateEquipSYNC.Encode())
	}
}

type FightReportPhaseResultToFightRouter struct {
	net.BaseRouter
}

func (p *FightReportPhaseResultToFightRouter) Handle(request iface.IRequest) {
	logrus.Info("***************************  多人对战阶段提交  ***************************")
	roleID, err := request.GetConnection().GetProperty("roleID")
	if err != nil {
		logrus.Error("GetProperty error:", err)
		return
	}
	player := core.WorldMgrObj.GetPlayerByPID(roleID.(int64))

	var cFightReportPhaseResultToFight = protocols.C_Fight_Report_Phase_Result_To_Fight{}
	cFightReportPhaseResultToFight.Decode(bytes.NewBuffer(request.GetData()), player.Key)
	logrus.Infof("%#v", cFightReportPhaseResultToFight)

	if cFightReportPhaseResultToFight.ReportData.FightType == 1 {
		var sFightReportPhaseResultToFight = protocols.S_Fight_Report_Phase_Result_To_Fight{Errorcode: 0}
		request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_Report_Phase_Result_To_Fight, sFightReportPhaseResultToFight.Encode())
	} else if cFightReportPhaseResultToFight.ReportData.FightType == 2 {
		// 合作

		var sFightReportPhaseResultToFight = protocols.S_Fight_Report_Phase_Result_To_Fight{Errorcode: 0}
		request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_Report_Phase_Result_To_Fight, sFightReportPhaseResultToFight.Encode())
	} else if cFightReportPhaseResultToFight.ReportData.FightType == 10 {

		var sFightReportPhaseResultToFight = protocols.S_Fight_Report_Phase_Result_To_Fight{Errorcode: 0}
		request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_Report_Phase_Result_To_Fight, sFightReportPhaseResultToFight.Encode())

	} else if cFightReportPhaseResultToFight.ReportData.FightType == 12 {
		// 寒冰堡

		var sFightReportPhaseResultToFight = protocols.S_Fight_Report_Phase_Result_To_Fight{Errorcode: 0}
		request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_Report_Phase_Result_To_Fight, sFightReportPhaseResultToFight.Encode())
	} else if cFightReportPhaseResultToFight.ReportData.FightType == 15 {
		// 机械迷城

		var sFightReportResultToFight = protocols.S_Fight_Report_Result_To_Fight{Errorcode: 0}
		request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_Report_Phase_Result_To_Fight, sFightReportResultToFight.Encode())
	}
}

type FightReportResultToFightRouter struct {
	net.BaseRouter
}

func (p *FightReportResultToFightRouter) Handle(request iface.IRequest) {
	logrus.Info("*******************************多人战斗结束提交*******************************")
	roleID, err := request.GetConnection().GetProperty("roleID")
	if err != nil {
		logrus.Error("GetProperty error:", err)
		return
	}
	player := core.WorldMgrObj.GetPlayerByPID(roleID.(int64))

	var cFightReportResultToFight = protocols.C_Fight_Report_Result_To_Fight{}
	cFightReportResultToFight.Decode(bytes.NewBuffer(request.GetData()), player.Key)
	logrus.Infof("%#v", cFightReportResultToFight)

	if cFightReportResultToFight.ReportData.FightType == 1 {
		// 角色对战
	} else if cFightReportResultToFight.ReportData.FightType == 2 {
		// 合作
		var sFightReportResultToFight = protocols.S_Fight_Report_Result_To_Fight{Errorcode: 0}
		request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_Report_Result_To_Fight, sFightReportResultToFight.Encode())
	} else if cFightReportResultToFight.ReportData.FightType == 10 {
		// 大航海
	} else if cFightReportResultToFight.ReportData.FightType == 12 {
		// 寒冰堡
		var sFightReportResultToFight = protocols.S_Fight_Report_Result_To_Fight{Errorcode: 0}
		request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_Report_Result_To_Fight, sFightReportResultToFight.Encode())
	} else if cFightReportResultToFight.ReportData.FightType == 15 {
		// 机械迷城
		var sFightReportResultToFight = protocols.S_Fight_Report_Result_To_Fight{Errorcode: 0}
		request.GetConnection().SendMessage(request.GetMsgType(), protocols.P_Fight_Report_Result_To_Fight, sFightReportResultToFight.Encode())
	}

	cNetworkFightToClientRoleFightBalance := protocols.C_Network_Fight_To_Logic_Role_FightBalance{
		RoleID: player.PID,
		SRoleFightBalance: protocols.S_Role_FightBalance{
			Type: cFightReportResultToFight.ReportData.FightType,
			BWin: cFightReportResultToFight.ReportData.IsWin,
			Roles: map[int64]protocols.T_FightBalance_Role{
				player.PID: {
					RoleAbstract: cFightReportResultToFight.ReportData.FightRoleInfo[player.PID].RoleAbstract,
				},
				player.FightInfo.RoleID: {
					RoleAbstract: cFightReportResultToFight.ReportData.FightRoleInfo[player.FightInfo.RoleID].RoleAbstract,
				},
			},
			Round:  cFightReportResultToFight.ReportData.FightRoleInfo[player.PID].Round,
			Battle: map[int32]protocols.T_FightBalance_Battle{},
			Coopration: map[int32]protocols.T_FightBalance_CoopRation{
				0: {
					Prize: []protocols.T_Reward{
						{DropType: 1, DropID: 1001, DropNum: 1950},
						{DropType: 1, DropID: 1002, DropNum: 105},
						{DropType: 2, DropID: 28, DropNum: 33},
						{DropType: 2, DropID: 9, DropNum: 5},
						{DropType: 2, DropID: 24, DropNum: 1},
						{DropType: 1, DropID: 1006, DropNum: 25},
						{DropType: 1, DropID: 1008, DropNum: 25},
						{DropType: 2, DropID: 18, DropNum: 1},
					},
					Extraprize: []protocols.T_Reward{},
				},
			},
			RandomArena:        map[int64]protocols.T_FightBalance_RandomArena{},
			GoldenLeague:       map[int32]protocols.T_FightBlance_GoldenLeague{},
			ActivityCoopration: map[int32]protocols.T_FightBalance_Activity_CoopRation{},
			ExtraData:          map[int32]protocols.T_FightBalance_ExtraData{},
		},
	}
	core.WorldMgrObj.SendMsg(protocols.P_Network_Fight_To_Logic, cNetworkFightToClientRoleFightBalance.Encode())
	// 对战结束，关闭连接
	request.GetConnection().Close()
}
