package db

import (
	"context"
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DbManager 全局的dbManager，方便controller那边操作数据库（增删改查）
var DbManager dbManager

func init() {
	// 初始化dbManager
	DbManager.InitDatabase()
}

// dbManager DB的管理器结构体
type dbManager struct {
	DBClient             *mongo.Client
	TFJLDatabase         *mongo.Database
	RoleCollection       *mongo.Collection // 角色
	FightItemsCollection *mongo.Collection // 对战items
}

// InitDatabase ...
func (manager *dbManager) InitDatabase() {
	credential := options.Credential{
		AuthSource: "admin",
		Username:   "",
		Password:   "",
	}
	// 设置客户端连接配置
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017").SetAuth(credential)
	// clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	clientOptions.SetMinPoolSize(10)
	clientOptions.SetMaxPoolSize(20)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 连接到MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// 检查连接
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	logrus.Info("Connected to MongoDB!")

	manager.DBClient = client
	manager.TFJLDatabase = client.Database("tfjl")
	logrus.Info("Connected to database: tfjl.")
}

func (manager *dbManager) CloseDB() {
	logrus.Info("closing db client connection...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := manager.DBClient.Disconnect(ctx); err != nil {
		panic(err)
	}
	logrus.Info("closed db client connection")
}
