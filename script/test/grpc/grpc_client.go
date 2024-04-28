package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	pb "github.com/1055373165/ggcache/api/groupcachepb"
	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/internal/middleware/etcd/discovery/discovery3"
	"github.com/1055373165/ggcache/internal/pkg/student/dao"
	"github.com/1055373165/ggcache/utils/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	ErrRPCCallNotFound  = "rpc error: code = Unknown desc = record not found"
	MaxRetries          = 3
	InitialRetryWaitSec = 1
)

const (
	NotFoundStatus Status = iota // 说明服务器没有查询到指定名字学生的分数
	ErrorStatus                  // 说明服务器出现问题
)

type Status int

func main() {
	config.InitConfig()
	dao.InitDB()

	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		panic(err)
	}

	// 服务发现（直接根据服务名字获取与服务的虚拟端连接）
	conn, err := discovery3.Discovery(cli, config.Conf.Services["groupcache"].Name)
	if err != nil {
		panic(err)
	}
	logger.LogrusObj.Debug("Discovery continue")
	client_stub := pb.NewGroupCacheClient(conn)

	// 为了模拟实际查询情况，将 names 数组中存在的名字和不存在的名字的分部比例设置为 10 : 1
	names := []string{"王五", "张三", "李四", "王二", "不存在", "赵六", "李奇"}
	for i := 0; i < 10; i++ {
		names = append(names, fmt.Sprintf("不存在%d", i))
	}
	for i := 0; i < 100; i++ {
		names = append(names, fmt.Sprintf("%d", i))
	}
	for i := 0; i < 100; i++ {
		names = append(names, fmt.Sprintf("%d", i))
	}
	for i := 0; i < 100; i++ {
		names = append(names, fmt.Sprintf("%d", i))
	}
	for i := 0; i < 100; i++ {
		names = append(names, fmt.Sprintf("%d", i))
	}
	// 打散
	rand.Shuffle(len(names), func(i, j int) {
		names[i], names[j] = names[j], names[i]
	})

	for {
		for _, name := range names {
			searchFunc := func() (*pb.GetResponse, error) {
				return client_stub.Get(context.Background(), &pb.GetRequest{
					Group: "scores",
					Key:   name,
				})
			}

			for i := 0; i < MaxRetries; i++ { // 重试机制
				resp, err := searchFunc()
				if err != nil {
					if ErrorHandle(err) == NotFoundStatus {
						logger.LogrusObj.Warnf("查询不到学生 %s 的成绩", name)
						break
					}
					logger.LogrusObj.Errorf("本次查询学生 %s 分数的 rpc 调用出现故障，重试次数 %d", name, i+1)
					waitTime := time.Duration(InitialRetryWaitSec*(1<<uint(i))) * time.Second // 退避算法
					time.Sleep(waitTime)
				} else {
					logger.LogrusObj.Infof("rpc 调用成功, 学生 %s 的成绩为 %s", name, string(resp.Value))
					break
				}
			}
		}
	}
}

func ErrorHandle(err error) Status {
	if err.Error() == ErrRPCCallNotFound {
		return NotFoundStatus
	}
	return ErrorStatus
}

// 第一次重试等待 1s
// 第二次重试等待 2s
// 第三次重试等待 4s
func backoff(retry int) int {
	return int(math.Pow(float64(2), float64(retry)))
}

// func main() {
// 	config.InitConfig()
// 	dao.InitDB()

// 	cli, err := clientv3.New(config.DefaultEtcdConfig)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// 服务发现（直接根据服务名字获取与服务的虚拟端连接）
// 	conn, err := discovery3.Discovery(cli, config.Conf.Services["groupcache"].Name)
// 	if err != nil {
// 		panic(err)
// 	}
// 	logger.LogrusObj.Debug("Discovery continue")
// 	client_stub := pb.NewGroupCacheClient(conn)

// 	// 为了模拟实际查询情况，将 names 数组中存在的名字和不存在的名字的分部比例设置为 10 : 1
// 	// 然后为了尽可能随机，使用洗牌算法将数组内所有学生姓名打散
// 	// 存在的名字中文 1000 个，英文 1000 个（已经存到数据库中了，测试集在主程序下方提供）
// 	// 查询请求使用 100 个已经存在中文名和 100 个已经存在的英文名
// 	// 使用不存在名字中文 10 个，英文 10 个

// 	unExistChineseNames := dao.GenerateChineseNames(10)
// 	unExistEnglishNames := dao.GenerateEnglishNames(10)

// 	dbExistChineseNames := dao.GetGenerateChineseNames()
// 	dbExistEnglishNames := dao.GetGenerateEnglishNames()

// 	names := make([]string, 0, 20+len(*dbExistChineseNames)+len(*dbExistEnglishNames))
// 	names = append(names, unExistChineseNames...)
// 	names = append(names, unExistEnglishNames...)
// 	names = append(names, *dbExistChineseNames...)
// 	names = append(names, *dbExistEnglishNames...)

// 	// rand.Shuffle(len(names), func(i, j int) {
// 	// 	names[i], names[j] = names[j], names[i]
// 	// })

// 	for {
// 		for _, name := range names {
// 			searchFunc := func() (*pb.GetResponse, error) {
// 				return client_stub.Get(context.Background(), &pb.GetRequest{
// 					Group: "scores",
// 					Key:   name,
// 				})
// 			}

// 			for i := 0; i < MaxRetries; i++ {
// 				resp, err := searchFunc()
// 				if err != nil {
// 					if ErrorHandle(err) == NotFoundStatus {
// 						logger.LogrusObj.Warnf("查询不到学生 %s 的成绩", name)
// 						break
// 					}
// 					logger.LogrusObj.Errorf("本次查询学生 %s 分数的 rpc 调用出现故障，重试次数 %d", name, i+1)
// 					time.Sleep(time.Second * time.Duration(backoff(i))) // 退避算法
// 				} else {
// 					logger.LogrusObj.Infof("rpc 调用成功, 学生 %s 的成绩为 %s", name, string(resp.Value))
// 					break
// 				}
// 			}
// 		}
// 	}
// }

/* 数据库中英文名称的所有测试数据
englishNames := []string{
	"Ella Robinson", "Alexander Williams", "James Franklin",
	"David Miller", "Matthew Jones", "Emma Hill", "John Smith",
	"David Lopez", "Daniel Green", "Chloe Scott", "Joseph Clark",
	"Olivia Stewart", "Olivia Taylor", "John Young", "Samuel Davis",
	"Isabella Hill", "Emily Baker", "Ava Jenkins", "Grace Adams",
	"Samuel Morris", "Joseph Bell", "Zoe Howard", "John Anderson",
	"Chloe Miller", "Samuel Collins", "Sophia Sanchez", "Joseph Scott",
	"Samuel Davis", "Isabella Walker", "Christopher Jenkins", "Ava Anderson",
	"Sophia Wilson", "William White", "Joseph Adams", "Benjamin Lopez",
	"Emma Washington", "Emma Harris", "Zoe King", "Sophia Adams",
	"David Johnson", "Sophia Thompson", "Jane Robinson", "Matthew Franklin",
	"Lily Evans", "Grace Smith", "Lily Brown", "Madison Washington",
	"Emily Franklin", "James Adams", "Chloe Nelson", "Mia Evans",
	"Olivia Harris", "Matthew Jefferson", "Ella Phillips", "Ava Thompson",
	"Ella Anderson", "Mia Johnson", "Mia Miller", "Lily Hall",
	"Hannah Jackson", "Hannah Harris", "Benjamin Bell", "Grace Roberts",
	"Daniel Garcia", "Jane Campbell", "Ethan Nelson", "Hannah Green",
	"Sophia Nelson", "Ethan Martinez", "Benjamin Franklin", "Olivia Campbell",
	"Mia Morgan", "Matthew Franklin", "Grace Jenkins", "Samuel Jones",
	"Ava Miller", "Emma Thompson", "David Robinson", "Noah Morris",
	"David Phillips", "Christopher Smith", "Matthew Martinez", "Madison Sanchez",
	"Joseph Walker", "Samuel Walker", "Madison Walker", "Samuel Campbell",
	"Christopher Jefferson", "Emma Williams", "Ethan Campbell", "Christopher Jenkins",
	"Samuel Thompson", "Noah Hall", "Olivia Green", "Grace Jenkins",
	"Ethan Baker", "William Collins", "Alexander Flores", "Emma Moore",
	"Benjamin Perry", "Noah Williams", "Daniel Smith", "Daniel Davis",
	"Sophia Baker", "Chloe Parker", "Daniel Phillips", "Emma Miller",
	"Ella Powell", "Chloe Taylor", "Lily Jones", "Noah Martin",
	"Ethan Campbell", "Zoe Flores", "Matthew Young", "Madison Jackson",
	"Zoe Harris", "James Evans", "William Baker", "Samuel Jackson",
	"Mia Howard", "David Evans", "Ella Smith", "Emma Hill",
	"Emma Franklin", "Samuel Harris", "Madison Moore", "Ethan Harris",
	"Ella Jackson", "Emma Rodriguez", "Sophia Garcia", "Chloe Franklin",
	"John Jones", "Benjamin Wilson", "Daniel Edwards", "Grace Young",
	"Sophia Scott", "Olivia Baker", "Madison Washington", "Ava Phillips",
	"David Hill", "Benjamin Thompson", "Daniel Thompson", "Daniel Williams",
	"Benjamin Garcia", "Grace Sanchez", "Samuel Garcia", "Isabella Campbell",
	"Chloe Baker", "Joseph Rodriguez", "Madison Anderson", "Samuel Campbell",
	"Noah Young", "Ella Parker", "Madison Campbell", "Zoe Robinson",
	"Emma Collins", "Emily Wright", "Ethan Edwards", "Ella Hill",
	"Zoe Smith", "Noah Jefferson", "Zoe Smith", "Mia Baker",
	"Jacob Roberts", "Samuel Green", "William Collins", "Noah Walker",
	"Emma Nelson", "Isabella Taylor", "David Powell", "Joseph Hill",
	"Joseph Morris", "Noah Roberts", "Sophia Campbell", "Zoe Jefferson",
	"Noah Young", "Christopher Hall", "Jane Stewart", "Matthew Parker",
	"Jacob Moore", "Daniel Hill", "Zoe White", "Ethan Collins",
	"Ethan Perry", "Jane Martin", "Sophia Perry", "Madison Jackson",
	"John Roberts", "Ethan Miller", "Matthew Collins", "Samuel Franklin",
	"Sophia Baker", "Michael Scott", "Lily Hill", "Samuel Thompson",
	"Hannah Franklin", "Ella Johnson", "Emma Thompson", "Grace Bell",
	"Emily Stewart", "Ethan Clark", "James Baker", "David Collins",
	"William Wilson", "Madison Hill", "Hannah Brown", "Christopher Garcia",
	"Lily Turner", "Zoe Scott", "Jane Roberts", "Hannah Jefferson",
	"Alexander Collins", "William Hall", "Grace Collins", "Olivia Adams",
	"Ella White", "Mia Stewart", "Matthew Scott", "Isabella Scott",
	"Ella Anderson", "Hannah Collins", "Chloe Morris", "Noah Green",
	"William Evans", "Jane Martinez", "Samuel Taylor", "Grace Campbell",
	"John Moore", "Madison Young", "Daniel White", "Jacob Jefferson",
	"Ava Turner", "Madison Robinson", "John Moore", "Madison Anderson",
	"Daniel Powell", "Sophia Clark", "Madison Morris", "Ella Taylor",
	"James Moore", "William Perry", "Ethan White", "Matthew Flores",
	"Benjamin Washington", "Chloe Morris", "William Nelson", "Lily Howard",
	"Jacob Evans", "Benjamin Lopez", "Ava Evans", "Emily Green",
	"Chloe Flores", "Madison Parker", "Michael Garcia", "Jane Thomas",
	"David Thomas", "Jacob Roberts", "Isabella Evans", "Daniel Turner",
	"Ethan King", "Daniel Taylor", "Matthew Young", "Jane Edwards",
	"Isabella Brown", "Jacob Martin", "David Roberts", "Isabella Nelson",
	"Isabella Taylor", "David Flores", "Isabella Collins", "Jacob Jefferson",
	"Benjamin Taylor", "Benjamin Baker", "Chloe Smith", "Lily Collins",
	"Emily Phillips", "John Perry", "Matthew Baker", "Noah Bell",
	"Lily King", "Jane Jenkins", "Jane Young", "Matthew Johnson",
	"James Smith", "Joseph Thomas", "Jacob Walker", "Ava Harris",
	"David Roberts", "Matthew Parker", "Ava Turner", "Grace Adams",
	"Hannah Evans", "Chloe Johnson", "Jane Morgan", "James Edwards",
	"Noah Jackson", "Christopher Hall", "Zoe Johnson", "Ava Hall",
	"John Williams", "Benjamin Miller", "Isabella Johnson", "Michael Franklin",
	"Noah Smith", "John Scott", "Samuel Rodriguez", "Matthew Baker",
	"Grace Stewart", "Olivia Edwards", "Benjamin Nelson", "Daniel Baker",
	"Madison Walker", "Mia Jones", "Joseph Johnson", "Isabella Moore",
	"Joseph Baker", "Mia Green", "Mia Turner", "Madison Lopez",
	"Matthew Campbell", "Ava Clark", "Noah Lopez", "David Rodriguez",
	"James Stewart", "Jacob Perry", "Jacob Brown", "Lily Davis",
	"Sophia Brown", "David Collins", "Olivia Martinez", "Isabella Wilson",
	"Jacob Wright", "Emily Turner", "Ava King", "Zoe Phillips",
	"Samuel Morgan", "Hannah Thompson", "Ethan Williams", "Christopher Robinson",
	"Zoe Roberts", "Ava Jenkins", "Ethan Perry", "Hannah Jefferson",
	"Jane Taylor", "Alexander Hill", "Grace King", "Samuel Moore",
	"Zoe Franklin", "Madison Sanchez", "Matthew Young", "Joseph Evans",
	"Matthew Evans", "Grace Anderson", "Jacob Stewart", "Christopher Brown",
	"John Powell", "Daniel Powell", "Benjamin Green", "Ella Davis",
	"Daniel Harris", "Sophia Martinez", "William Martinez", "John Jones",
	"James Morgan", "Hannah Johnson", "Zoe Jefferson", "James Rodriguez",
	"James Phillips", "John Baker", "Lily Morris", "David Jefferson",
	"John Martin", "Noah Lopez", "Jane Anderson", "David Green",
	"Ella Franklin", "Ella Morris", "Hannah Adams", "Michael Wilson",
	"Alexander Scott", "Alexander Anderson", "Olivia White", "Chloe Roberts",
	"Hannah Campbell", "Zoe Wright", "John White", "Madison King",
	"Benjamin Hill", "Emily Campbell", "Sophia Thompson", "Zoe Adams",
	"Michael Taylor", "Samuel Adams", "Lily Smith", "Grace Edwards",
	"Chloe Washington", "Joseph Jackson", "Daniel Adams", "Olivia White",
	"Joseph Harris", "David Williams", "Matthew Collins", "Ethan Stewart",
	"Samuel Turner", "Ethan Turner", "Sophia White", "John Lopez",
	"Emma Hall", "James Stewart", "Michael Martinez", "Ella Johnson",
	"Daniel Edwards", "Mia Campbell", "William Wright", "Ethan Turner",
	"Ella Thompson", "Chloe Wright", "Alexander Wilson", "Sophia Scott",
	"Joseph White", "Daniel Rodriguez", "John Evans", "Ella Smith",
	"Joseph Evans", "Ethan Hall", "Sophia Harris"]
*/
