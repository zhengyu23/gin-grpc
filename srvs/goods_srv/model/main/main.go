package main

import (
	"context"
	"crypto/md5" // md5 的包
	"encoding/hex"
	"io"
	"log"
	"os"
	"strconv"
	"time"
	"xrUncle/srvs/goods_srv/global"
	"xrUncle/srvs/goods_srv/model"

	"github.com/olivere/elastic/v7"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

/*
	连接数据库
	md5 + 盐值加密
*/

func genMd5(code string) string {
	Md5 := md5.New()
	_, _ = io.WriteString(Md5, code)
	return hex.EncodeToString(Md5.Sum(nil))

}

func main() {
	////连接数据库
	//dsn := "root:root@tcp(169.254.185.5:3306)/chenbing_goods_srv?charset=utf8mb4&parseTime=True&loc=Local"
	//
	//// 设置全局的logger，它会在我们执行每个sql语句的时候打印一行sql
	//newLogger := logger.New(
	//	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
	//	logger.Config{
	//		SlowThreshold:              time.Second,   // Slow SQL threshold
	//		LogLevel:                   logger.Info, // Log level
	//		IgnoreRecordNotFoundError: true,           // Ignore ErrRecordNotFound error for logger
	//		Colorful:                  true,          // Disable color
	//	},
	//)
	//
	//// 全局模式
	//var err error
	//db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
	//	NamingStrategy: schema.NamingStrategy{
	//		SingularTable: true,	// 表名不加 true
	//	},
	//	Logger: newLogger,
	//
	//})
	//if err != nil {
	//	panic("连接失败")
	//}
	////_ = db.AutoMigrate(&model.User{})
	//
	//// 导入表数据
	//options := &password.Options{16, 100, 32, sha512.New}
	//salt, encodedPwd := password.Encode("admin123", options)
	//newPassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)
	//
	//for i := 0; i < 10; i++ {
	//	user := model.User{
	//		NickName: fmt.Sprintf("Anthony%d",i),
	//		Mobile: fmt.Sprintf("1368290028%d",i),
	//		Password: newPassword,	// admin123-md5
	//	}
	//	db.Save(&user)	// 会修改ID，所以要加入指针
	//}
	/*
		fmt.Println(genMd5("123456"))
		// 暴力破解 生成 123456 11111 的彩虹表 -> 所以采用盐值加密
		// 米饭1 + 米饭2  ->  撒盐  ->  完全不同的味道，无从猜测
		// salt ： 随机字符串 +用户密码

		// Using the default options
		salt, encodedPwd := password.Encode("generic password", nil)
		fmt.Println(salt)
		fmt.Println(encodedPwd)
		check := password.Verify("generic password", salt, encodedPwd, nil)
		fmt.Println(check) // true

		// Using custom options		修改 salt长度、key长度
		options := &password.Options{16, 100, 32, sha512.New}
		salt, encodedPwd = password.Encode("generic password", options)
		// 添加盐值的密码	用 $ 分割
		newPassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)
		fmt.Println(len(newPassword))
		fmt.Println(newPassword)
		passwordInfo := strings.Split(newPassword, "$")	// 通过 $分割
		fmt.Println(passwordInfo)
		check = password.Verify("generic password", salt, encodedPwd, options)
		fmt.Println(check) // true
	*/
	Mysql2Es()
}

// Mysql2Es 数据库商品存储到ES
func Mysql2Es() {
	dsn := "root:root@tcp(192.168.31.142:3306)/chenbing_goods_srv?charset=utf8mb4&parseTime=True&loc=Local"

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // 禁用彩色打印
		},
	)

	// 全局模式
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: newLogger,
	})
	if err != nil {
		panic(err)
	}

	host := "http://192.168.31.142:9200"
	logger := log.New(os.Stdout, "chenbing", log.LstdFlags)
	global.EsClient, err = elastic.NewClient(elastic.SetURL(host), elastic.SetSniff(false),
		elastic.SetTraceLog(logger))
	if err != nil {
		panic(err)
	}

	var goods []model.Goods
	db.Find(&goods)
	for _, g := range goods {
		esModel := model.EsGoods{
			ID:          g.ID,
			CategoryID:  g.CategoryID,
			BrandsID:    g.BrandsID,
			OnSale:      g.OnSale,
			ShipFree:    g.ShipFree,
			IsNew:       g.IsNew,
			IsHot:       g.IsHot,
			Name:        g.Name,
			ClickNum:    g.ClickNum,
			SoldNum:     g.SoldNum,
			FavNum:      g.FavNum,
			MarketPrice: g.MarketPrice,
			GoodsBrief:  g.GoodsBrief,
			ShopPrice:   g.ShopPrice,
		}

		_, err = global.EsClient.Index().Index(esModel.GetIndexName()).BodyJson(esModel).Id(strconv.Itoa(int(g.ID))).Do(context.Background())
		if err != nil {
			panic(err)
		}
		//强调一下 一定要将docker启动es的java_ops的内存设置大一些 否则运行过程中会出现 bad request错误
	}
}
