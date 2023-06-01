package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sync"
	"xrUncle/srvs/inventory_srv/global"
	"xrUncle/srvs/inventory_srv/model"
	"xrUncle/srvs/inventory_srv/proto"
)

type InventoryServer struct {
	proto.UnimplementedInventoryServer
}

// 设置库存
func (*InventoryServer) SetInv(ctx context.Context, req *proto.GoodsInvInfo) (*emptypb.Empty, error) {
	//设置库存， 如果我要更新库存
	var inv model.Inventory
	global.DB.Where(&model.Inventory{Goods: req.GoodsId}).First(&inv, req.GoodsId)
	inv.Goods = req.GoodsId
	inv.Stocks = req.Num

	global.DB.Save(&inv)
	return &emptypb.Empty{}, nil
}

// 商品库存详情
func (*InventoryServer) InvDetail(ctx context.Context, req *proto.GoodsInvInfo) (*proto.GoodsInvInfo, error) {
	var inv model.Inventory
	if result := global.DB.Where(&model.Inventory{Goods: req.GoodsId}).First(&inv); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "没有库存信息")
	}
	return &proto.GoodsInvInfo{
		GoodsId: inv.Goods,
		Num:     inv.Stocks,
	}, nil
}

// 分布式锁 - 基于 redsync，解决互斥性(setnx)、死锁(expiry)、安全性(值认证)
// 库存扣减
func (*InventoryServer) Sell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	client := goredislib.NewClient(&goredislib.Options{
		Addr: "192.168.31.142:6379", // docker容器中的redis
		// TODO 修改成nacos配置
	})
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)
	rs := redsync.New(pool)

	tx := global.DB.Begin()

	// sellDetail 库存扣减历史表的写入
	sellDetail := model.StockSellDetail{
		OrderSn: req.OrderSn,
		Status:  1, // 1 表示已扣减
	}
	var details []model.GoodsDetail // GoodsDetail List

	for _, goodInfo := range req.GoodsInfo {
		var inv model.Inventory

		// GoodsDetail List
		details = append(details, model.GoodsDetail{
			Goods: goodInfo.GoodsId,
			Num:   goodInfo.Num,
		})

		mutex := rs.NewMutex(fmt.Sprintf("goods_%d", goodInfo.GoodsId))
		if err := mutex.Lock(); err != nil {
			return nil, status.Errorf(codes.Internal, "获取redis分布式锁异常")
		}

		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		}
		//判断库存是否充足
		if inv.Stocks < goodInfo.Num {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}
		inv.Stocks -= goodInfo.Num
		tx.Save(&inv)

		if ok, err := mutex.Unlock(); !ok || err != nil {
			return nil, status.Errorf(codes.Internal, "释放redis分布式锁异常")
		}
	}

	// 生成库存扣减历史表 sellDetail
	sellDetail.Detail = details
	if result := tx.Create(&sellDetail); result.RowsAffected == 0 {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "保存库存扣减历史失败")
	}

	tx.Commit() // 需要自己手动提交操作
	return &emptypb.Empty{}, nil
}

// 分布式锁 - 乐观锁 ③
func (*InventoryServer) SellOptimisticLock(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	tx := global.DB.Begin()
	for _, goodInfo := range req.GoodsInfo {
		var inv model.Inventory
		for {
			if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
				tx.Rollback() //回滚之前的操作
				return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
			}
			//判断库存是否充足
			if inv.Stocks < goodInfo.Num {
				tx.Rollback() //回滚之前的操作
				return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
			}
			inv.Stocks -= goodInfo.Num
			// UPDATE inventory SET stocks=stocks-1, version=version+1 WHERE goods=goods and version=version
			// 这种写法有瑕疵， 为什么？
			// 零值 对于int类型来说，默认值为 0，其会被 gorm忽略
			// 加上Select方法固定，可以包含零值
			if result := tx.Model(&model.Inventory{}).Select("Stocks", "Version").Where("goods = ? and version = ?",
				goodInfo.GoodsId, inv.Version).Updates(model.Inventory{Stocks: inv.Stocks, Version: inv.Version + 1}); result.RowsAffected == 0 {
				zap.S().Info("库存扣减失败")
			} else {
				break
			}
		}
	}
	tx.Commit() // 需要自己手动提交操作
	//m.Unlock()
	return &emptypb.Empty{}, nil
}

// 分布式锁 - 悲观锁 ②
func (*InventoryServer) SellPessimisticLock(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//扣减库存， 本地事务 [1:10,  2:5, 3: 20]
	//数据库基本的一个应用场景：数据库事务
	//并发情况之下 可能会出现超卖 1
	tx := global.DB.Begin()
	for _, goodInfo := range req.GoodsInfo {
		var inv model.Inventory
		if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		}
		//判断库存是否充足
		if inv.Stocks < goodInfo.Num {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}
		//扣减， 会出现数据不一致的问题 - 锁，分布式锁
		inv.Stocks -= goodInfo.Num
		tx.Save(&inv)
	}
	tx.Commit() // 需要自己手动提交操作
	return &emptypb.Empty{}, nil
}

// 基于 GO自带锁 ①
var m sync.Mutex

func (*InventoryServer) SellGoMutex(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//扣减库存， 本地事务 [1:10,  2:5, 3: 20]
	//数据库基本的一个应用场景：数据库事务
	//并发情况之下 可能会出现超卖 1
	tx := global.DB.Begin()
	m.Lock() //这把锁有问题吗？ 假设有 10w的并发，请求的并不是同一间商品
	for _, goodInfo := range req.GoodsInfo {
		var inv model.Inventory
		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			//if result := tx.Clauses(clause.Locking{Strength:"UPDATE"}).Where(&model.Inventory{Goods:goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		}
		//判断库存是否充足
		if inv.Stocks < goodInfo.Num {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}
		//扣减， 会出现数据不一致的问题 - 锁，分布式锁
		inv.Stocks -= goodInfo.Num
		tx.Save(&inv)
	}
	tx.Commit() // 需要自己手动提交操作
	m.Unlock()
	return &emptypb.Empty{}, nil
}

// 库存归还
func (*InventoryServer) Reback(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//库存归还： 1：订单超时归还 2. 订单创建失败，归还之前扣减的库存 3. 手动归还
	tx := global.DB.Begin()
	for _, goodInfo := range req.GoodsInfo {
		var inv model.Inventory
		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback()
			return nil, status.Errorf(codes.NotFound, "没有库存信息")
		}
		//扣减， 会出现数据不一致的问题 - 锁，分布式锁
		inv.Stocks += goodInfo.Num
		tx.Save(&inv)
	}
	tx.Commit() // 需要自己手动提交操作
	return &emptypb.Empty{}, nil
}

func AutoReback(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	//既然是归还库存，那么我应该具体的知道每件商品应该归还多少， 但是有一个问题是什么？重复归还的问题
	//所以说这个接口应该确保幂等性， 你不能因为消息的重复发送导致一个订单的库存归还多次， 没有扣减的库存你别归还
	//如果确保这些都没有问题， 新建一张表， 这张表记录了详细的订单扣减细节，以及归还细节

	type OrderInfo struct {
		OrderSn string
	}
	// 1、拿到msgs中数据 拿到订单索引 - OrderSn
	for i := range msgs {
		var orderInfo OrderInfo
		if err := json.Unmarshal(msgs[i].Body, &orderInfo); err != nil {
			// 加入消息提取失败，认为msgs里为无用消息，废弃之
			return consumer.ConsumeSuccess, nil
		}
		// 2、将inv库存归还，3、设置sellDetail表中status为已归还（在事务中进行）
		tx := global.DB.Begin()
		var sellDetail model.StockSellDetail
		if result := tx.Model(&model.StockSellDetail{}).
			Where(&model.StockSellDetail{OrderSn: orderInfo.OrderSn}).First(&sellDetail); result.RowsAffected == 0 {
			// 说明已经归还过
			return consumer.ConsumeSuccess, nil
		}
		// 如果查询到历史 则逐个归还
		for _, orderGood := range sellDetail.Detail {
			if result := tx.Model(&model.Inventory{}).
				Where(&model.Inventory{Goods: orderGood.Goods}).
				Update("stocks", gorm.Expr("stocks+?", orderGood.Num)); result.RowsAffected == 0 {
				tx.Rollback()
				return consumer.ConsumeRetryLater, nil
			}
		}

		if result := tx.Model(&model.StockSellDetail{}).
			Where(&model.StockSellDetail{OrderSn: orderInfo.OrderSn}).
			Update("status", 2); result.RowsAffected == 0 {
			tx.Rollback()
			return consumer.ConsumeRetryLater, nil
		}
		tx.Commit()
		//return consumer.ConsumeSuccess,nil
	}
	return consumer.ConsumeSuccess, nil
}
