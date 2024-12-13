package dao

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	stuPb "github.com/1055373165/ggcache/api/studentpb"
	"github.com/1055373165/ggcache/internal/pkg/student/model"
	"github.com/1055373165/ggcache/utils/logger"
)

func migration() {
	if IsHasTable("student") {
		return
	}

	err := _db.Set("gorm:table_options", "charset=utf8mb4").
		AutoMigrate(
			&model.Student{},
		)

	if err != nil {
		logger.LogrusObj.Infoln("register table failed")
		os.Exit(0)
	}

	// InitilizeDB()
	logger.LogrusObj.Infoln("register table success")
}

func IsHasTable(tableName string) bool {
	return _db.Migrator().HasTable(tableName)
}

func InitilizeDB() {
	d := NewStudentDao(context.Background())

	names := []string{"王五", "张三", "李四", "王二", "赵六", "李奇"}

	for _, name := range names {
		d.CreateStudent(&stuPb.StudentRequest{
			Name:  name,
			Score: float32(rand.Int31n(10000)),
		})
	}

	for i := 0; i < 1000; i++ {
		d.CreateStudent(&stuPb.StudentRequest{
			Name:  fmt.Sprintf("%d", i),
			Score: float32(rand.Int31n(100)),
		})
	}

	logger.LogrusObj.Infoln("数据导入成功...")
}

func GenerateChineseNames(n int) []string {
	surnames := []string{"李", "王", "张", "刘", "陈", "杨", "赵", "黄", "周", "吴",
		"徐", "孙", "朱", "马", "胡", "郭", "林", "高", "罗", "梁",
		"宋", "郑", "谢", "韩", "唐", "冯", "于", "董", "萧", "程"}
	givenNames := []string{"的", "了", "在", "是", "我", "有", "和", "就", "不", "人",
		"都", "一", "一个", "上", "也", "很", "到", "说", "要", "去",
		"你", "会", "着", "没有", "看", "好", "自己", "这", "那", "还",
		"们", "大", "来", "为", "着", "里", "它", "她", "又", "什么",
		"把", "好像", "知道", "能", "可以", "觉得", "是的", "时候", "怎样", "没"}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	names := make([]string, n)
	for i := 0; i < n; i++ {
		surname := surnames[r.Intn(len(surnames))]
		givenName := givenNames[r.Intn(len(givenNames))]
		names[i] = fmt.Sprintf("%s%s", surname, givenName)
	}

	return names
}

func GenerateEnglishNames(n int) []string {
	firstNames := []string{
		"John", "Jane", "Michael", "Emma", "William", "Olivia",
		"James", "Ava", "David", "Isabella", "Benjamin", "Sophia",
		"Matthew", "Mia", "Ethan", "Chloe", "Daniel", "Zoe",
		"Jacob", "Emily", "Christopher", "Madison", "Joseph", "Ella",
		"Alexander", "Lily", "Samuel", "Grace", "Noah", "Hannah",
	}
	lastNames := []string{
		"Smith", "Johnson", "Williams", "Jones", "Brown", "Davis",
		"Miller", "Wilson", "Moore", "Taylor", "Anderson", "Thomas",
		"Jackson", "White", "Harris", "Martin", "Thompson", "Garcia",
		"Martinez", "Robinson", "Clark", "Rodriguez", "Wright", "Lopez",
		"Washington", "Jefferson", "Morgan", "Franklin", "Scott", "King",
		"Walker", "Bell", "Young", "Nelson", "Baker", "Hall",
		"Roberts", "Green", "Adams", "Evans", "Turner", "Phillips",
		"Parker", "Howard", "Collins", "Jenkins", "Perry", "Stewart",
		"Sanchez", "Morris", "Powell", "Hill", "Baker", "Campbell",
		"Flores", "Edwards",
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	names := make([]string, n)
	for i := 0; i < n; i++ {
		firstName := firstNames[r.Intn(len(firstNames))]
		lastName := lastNames[r.Intn(len(lastNames))]
		names[i] = firstName + " " + lastName
	}

	return names
}

// 自己从数据库中提取的
func GetGenerateEnglishNames() *[]string {
	return &[]string{
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
		"Benjamin Perry",
	}
}

func GetGenerateChineseNames() *[]string {
	return &[]string{
		"李说", "林它", "宋什么", "刘好像", "杨什么", "胡着", "郭看", "刘好", "萧她", "刘没",
		"赵你", "董没", "朱了", "王不", "陈一", "董可以", "罗来", "宋上", "于什么", "程为",
		"张没有", "于在", "郭时候", "李没", "郑是的", "张好", "韩它", "冯怎样", "罗没有", "王可以",
		"马你", "周就", "罗又", "吴有", "高着", "赵到", "李我", "梁里", "周可以", "林这",
		"董了", "黄看", "郭上", "罗和", "宋可以", "杨它", "宋你", "陈是的", "郭们", "周要",
		"李没有", "赵和", "程了", "梁觉得", "唐人", "马里", "张在", "宋为", "梁是", "萧怎样",
		"马了", "梁怎样", "董在", "董为", "杨是", "高在", "罗为", "韩一个", "黄着", "冯不",
		"徐自己", "郭上", "周可以", "黄好像", "萧有", "陈怎样", "谢来", "罗都", "赵好", "高着",
		"杨里", "陈觉得", "徐还", "冯知道", "郭着", "杨们", "徐大", "萧没有", "郭看", "黄一个",
		"孙这", "胡要", "冯你", "杨可以", "宋它", "冯不", "张把", "谢大", "程了", "郑还",
	}
}
