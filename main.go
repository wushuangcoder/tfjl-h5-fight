package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"tfjl-h5-fight/apis"
	"tfjl-h5-fight/configs"
	"tfjl-h5-fight/core"
	"tfjl-h5-fight/db"
	"tfjl-h5-fight/iface"
	"tfjl-h5-fight/net"
	"tfjl-h5-fight/protocols"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var (
	configFile string
)

func init() {
	logrus.SetOutput(os.Stdout)
	db.DbManager.SetRoleCollection("role")
	db.DbManager.SetFightItemsCollection("fight_items")
}

func initCmd() {
	flag.StringVar(&configFile, "config", "./config.json", "where load config json")
	flag.Parse()
}

func OnConnectionLost(conn iface.IConnection) {
	roleID, err := conn.GetProperty("roleID")
	if err != nil {
		logrus.Info("conn GetProperty roleID error:", err)
		return
	}
	player := core.WorldMgrObj.GetPlayerByPID(roleID.(int64))
	// 玩家对战服务器下线
	player.Offline()
	logrus.Info("========> Player roleID =", roleID, "offline... <========")
}

func main() {
	initCmd()
	if err := configs.LoadConfig(configFile); err != nil {
		fmt.Println("Load config json error:", err)
	}
	net.WSServer = net.NewServer()
	net.WSServer.SetOnConnStop(OnConnectionLost)
	// 对战-角色登录
	net.WSServer.AddRouter(protocols.P_Fight_Role_Login, &apis.FightRoleLoginRouter{})
	// 对战-进房加载准备
	net.WSServer.AddRouter(protocols.P_Fight_Loading_Ready, &apis.FightLoadingReadyRouter{})
	// 对战-同步银币
	net.WSServer.AddRouter(protocols.P_Fight_Silver_SYNC, &apis.FightSilverSYNCRouter{})
	// 对战-同步卡牌刷新次数
	net.WSServer.AddRouter(protocols.P_Fight_Refresh_Card_Count_SYNC, &apis.FightRefreshCardCountSYNCRouter{})
	// 对战-同步英雄属性
	net.WSServer.AddRouter(protocols.P_Fight_Hero_Attr_SYNC, &apis.FightHeroAttrSYNCRouter{})
	// 对战-血量同步
	net.WSServer.AddRouter(protocols.P_Fight_Monster_Blood_SYNC, &apis.FightMonsterBloodSYNCRouter{})
	// 对战-同步战车
	net.WSServer.AddRouter(protocols.P_Fight_Update_Car_SYNC, &apis.FightUpdateCarSYNCRouter{})
	// 对战-更新英雄
	net.WSServer.AddRouter(protocols.P_Fight_Update_Hero_SYNC, &apis.FightUpdateHeroSYNCRouter{})
	// 对战-出售英雄
	net.WSServer.AddRouter(protocols.P_Fight_Sell_Hero_SYNC, &apis.FightSellHeroSYNCRouter{})
	// 对战-英雄技能同步
	net.WSServer.AddRouter(protocols.P_Fight_Skill_SYNC, &apis.FightSkillSYNCRouter{})
	// 对战-操作装备同步
	net.WSServer.AddRouter(protocols.P_Fight_Operate_Equip_SYNC, &apis.FightOperateEquipSYNCRouter{})

	// 对战-提交每阶段的战斗结果到对战服务器
	net.WSServer.AddRouter(protocols.P_Fight_Report_Phase_Result_To_Fight, &apis.FightReportPhaseResultToFightRouter{})
	// 对战-提交结束的战斗结果到对战服务器
	net.WSServer.AddRouter(protocols.P_Fight_Report_Result_To_Fight, &apis.FightReportResultToFightRouter{})

	router := gin.Default()
	router.GET("/tfjlh5/fight/ws", net.WsHandler)
	bindAddress := fmt.Sprintf("%s:%d", configs.GConf.Ip, configs.GConf.Port)
	srv := &http.Server{
		Addr:    bindAddress,
		Handler: router,
	}

	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("listen: %s\n", err)
		}
	}()

	// 创建一个WebSocket客户端，用于连接主逻辑服务器，对战完将对战结果提交到主逻辑服务器，主逻辑服务器再将结果返回给游戏客户端
	headers := http.Header{}
	headers.Set("Authorization", "e756795a-1245-458f-ae1c-8f1e2ccf5e28")
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/tfjlh5/ws", headers)
	if err != nil {
		logrus.Error("websocket dial error:", err)
		return
	}
	core.WorldMgrObj.Client = conn
	defer core.WorldMgrObj.Client.Close()

	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal := <-quit
	log.Println("Shutdown Server ..., Signal:", signal)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 关闭gin的http服务器
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
	// 关闭数据库
	db.DbManager.CloseDB()
}
