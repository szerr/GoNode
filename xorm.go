package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"log"
	"time"
)

var engine *xorm.Engine

const DBNAME = "xorm" //测试用的数据库名

type GroupUser struct {
	User  `xorm:"extends"`
	Group `xorm:"extends"`
}

func (GroupUser) TableName() string {
	return "user"
}

type User struct { //如果和关键字冲突，需要加上单引号
	Id         int64
	Name       string `xorm:"varchar(32) index notnull 'user_name'"`
	Age        int
	Gender     string            `xorm:"varchar(8) default('男')"` //这里的中文 需要加个单引号
	Other      map[string]string `xorm:"json"`                    //复合数据类型可以用json存储
	CreateTime time.Time         `xorm:"created"`                 //创建时自动赋值
	DeleteTime time.Time         `xorm:"deleted"`                 //删除时自动赋值 and 不删除
	UpdateTime time.Time         `xorm:"updated"`                 //更新时自动赋值
	GroupId    int64             `xorm:"index"`
	Version    int               `xorm:"version"` //乐观锁用的字段
}

type Group struct {
	Id   int64
	Name string
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Ldate | log.Ltime)
}

func CheckErr(args ...interface{}) interface{} {
	var err error
	var redata interface{}
	if len(args) > 1 {
		err, _ = args[1].(error) //这里的nil断言成error会失败，所以不用管是否断言成功
		redata = args[0]
	} else {
		err, _ = args[0].(error)
	}
	if err != nil {
		log.Fatal(err)
	}
	return redata
}

func GetEngine() *xorm.Engine {
	var err error
	// 创建一个mysql engine
	engine, err = xorm.NewEngine("mysql", "root@/"+DBNAME)
	//这里的数据库连接串格式：[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	//例如 readonly:bulabula@tcp(127.0.0.1:3306)/test?charset=utf-8
	//engine是GoRoutine安全的
	//NewEngine的参数和sql.Open传入的参数相同，在用某个驱动前，可以看传入参数的文档:http://xorm.io/docs
	CheckErr(err)
	return engine
}

func TestDBConn(engine *xorm.Engine) {
	CheckErr(engine.Ping()) //第一次访问时开始链接数据库, Ping用来做数据库链接测试
	log.Println("db connect ok")
}

func Echo(args ...interface{}) {
	//给提示输出加上颜色
	print("\033[1;32m")
	log.Println(args...)
	print("\033[0m")
}

func Echor(args ...interface{}) {
	//给提示输出加上颜色
	print("\033[32m返回结果：")
	fmt.Println(args...)
	print("\033[0m")
}

func InitGroup(engine *xorm.Engine) {
	for _, name := range []string{"man", "woman"} {
		group := new(Group)
		has := CheckErr(engine.Where("Name=?", name).Get(group))
		if has == false {
			group.Name = name
			engine.Insert(group)
		}
	}
}

func InsertTest(engine *xorm.Engine) int64 { //插入数据的例子
	var all_insert_num int64
	other := make(map[string]string)
	other["bula"] = "didi"
	other["didi"] = "bula"
	group1 := new(Group)
	group2 := new(Group)
	CheckErr(engine.Where("name='man'").Get(group1))
	CheckErr(engine.Where("name='woman'").Get(group2))
	user1 := User{
		Name:    "bulabula",
		Age:     17,
		Gender:  "男",
		Other:   other,
		GroupId: group1.Id,
	}
	user2 := User{
		Name:    "didadida",
		Age:     18,
		Gender:  "女",
		Other:   other,
		GroupId: group2.Id,
	}
	insert_num, err := engine.Insert(&user1, &user2) //插两条
	CheckErr(err)
	all_insert_num += insert_num
	user2.Id += 1
	user2.Age = 19
	insert_num, err = engine.InsertOne(&user2) //插一条
	all_insert_num += insert_num
	CheckErr(err)
	return all_insert_num
}

func main() {
	engine := GetEngine() //这里还没有连接数据库
	TestDBConn(engine)
	defer engine.Close()                                      //可以这样关闭链接，一般情况下可以不关，程序关闭时会自动关闭链接
	engine.ShowSQL(true)                                      //打印日志到控制台
	engine.TZLocation, _ = time.LoadLocation("Asia/Shanghai") //默认是Local时区，这里改成其他时区
	Echo("同步数据库结构：")
	CheckErr(engine.Sync2(new(User)))
	CheckErr(engine.Sync2(new(Group)))
	InitGroup(engine) //初始化组数据

	user := new(User)
	users := []User{}

	Echo("先清掉所有数据：")
	Echor(CheckErr(engine.Query("delete from user")))

	//插入数据：
	Echo("插入测试")
	Echo("插入数量: ", InsertTest(engine))

	//查询数据：
	Echo("查询第一条数据：")
	Echor(CheckErr(engine.Get(user)), *user)
	Echo("根据struct中已有的数据查询")
	user = &User{Name: user.Name}
	Echor(CheckErr(engine.Get(user)), *user)
	Echo("Get 会返回 has 和 err， has表示是否存在，err表示错误，不管err是否为nil，has都有可能true或者false")

	Echo("查询所有数据：")
	Echor(CheckErr(engine.Find(&users)), users)
	Echo("Find也可以接收Map指针，map的key为Id")
	users_map := make(map[int64]User)
	Echor(CheckErr(engine.Find(&users_map)), users_map)
	Echo("Find 如果只选择单个字段，也可以用非结构体的Slice")
	var li []int64
	Echor(CheckErr(engine.Table("user").Cols("id").Find(&li)), li)

	Echo("条件查询：")
	user = new(User) //注意每次查询前需要new一个新struct，不然会根据struct的值查询
	Echor(CheckErr(engine.Where("age=? and gender=?", "18", "女").Get(user)), *user)
	Echo("也可以：")
	user = new(User)
	Echor(CheckErr(engine.Where("age=?", "18").And("gender=?", "女").Get(user)), *user)
	Echo("注意xorm log 里的sql是不一样的")
	user = new(User)
	Echo("还有一个：")
	Echor(CheckErr(engine.Where("age=?", "18").Or("gender=?", "女").Get(user)), *user)

	Echo("用struct的值做条件查询：")
	user = new(User)
	user.Id = 2
	Echor(CheckErr(engine.Get(user)), *user)

	user = new(User)
	Echo("排序取最小：")
	Echor(CheckErr(engine.Asc("age").Get(user)), *user)
	user = new(User)
	Echo("也可以：")
	Echor(CheckErr(engine.OrderBy("age").Get(user)), *user)
	user = new(User)
	Echo("排序取最大：")
	Echor(CheckErr(engine.Desc("age").Get(user)), *user)
	user = new(User)
	Echo("也可以：")
	Echor(CheckErr(engine.OrderBy("-age").Get(user)), *user)

	user = new(User)
	Echo("用主键查询：")
	Echor(CheckErr(engine.Id(1).Get(user)), *user)

	//Echo("复合主键？这里没有这种东西")
	//engine.Id(core.PK{1, "name"}).Get(&user)
	//这里的两个参数按照struct中pk标记字段出现的顺序赋值。

	user = new(User)
	CheckErr(engine.Desc("age").Get(user)) //下边需要这个id，所以就不
	Echo("指定字段：")
	user = new(User)
	Echor(CheckErr(engine.Select("user_name, age").Id(1).Get(user)), *user)

	Echo("直接跑sql, 这里可以不初始化struct")
	user = new(User)
	Echor(CheckErr(engine.Sql("select * from user where id=2").Get(user)), *user)

	Echo("某字段在一些值中：")
	user = new(User)
	users = []User{}
	Echor(CheckErr(engine.In("id", 1, 2, 3).Find(&users)), users)

	Echo("用Cols指定查询某些字段：")
	user = new(User)
	Echor(CheckErr(engine.Cols("age", "user_name").Get(user)), *user)
	Echo("用Cols指定更新某些字段：")
	user.Name = "haha"
	Echor(CheckErr(engine.Id(user.Id).Cols("user_name", "age").Update(user)), *user)

	Echo("指定操作所有字段：一般与Update配合使用，因为默认Update只更新非0，非”“，非bool的字段。这里只做个查询 = = ")
	Echor(CheckErr(engine.Where("id=?", user.Id).AllCols().Update(user)), *user)

	Echo("必须更新某些字段：让Update更新非0，非”“，非bool的字段。这个功能和Cols一样")
	Echor(CheckErr(engine.Id(user.Id).MustCols("age").Update(user)), *user)

	Echo("Update 也可以用map选择字段更新：返回更新的记录数")
	user = new(User)
	CheckErr(engine.Desc("age").Get(user)) //下边需要这个id，所以就不
	Echor(CheckErr(engine.Table(user).Id(user.Id).Cols("age").Update(map[string]interface{}{"age": 19, "Version":1})))

	Echo("这里User因为启用的乐观锁，所以每次Update Version都会+1，每次更新必须包含version原来的值。用map更新时乐观锁字段必须传golang中的字段名而不是表里的字段名。")
	Echo("")

	//再插一坨数据
	InsertTest(engine)
	InsertTest(engine)

	Echo("归类结果：")
	users = []User{}
	Echor(CheckErr(engine.Distinct("age", "user_name").Find(&users)), users)

	Echo("指定对哪个表做操作：可以用结构体")
	user = new(User)
	Echor(CheckErr(engine.Table(&user).Get(user)), user)
	Echo("或者表名：")
	user = new(User)
	Echor(CheckErr(engine.Table("user").Get(user)), user)

	Echo("限制获取的条目：获取两条，从第二条开始，第二个参数不传为0")
	users = []User{}
	Echor(CheckErr(engine.Where("age=17").Limit(3, 1).Find(&users)), users)

	Echo("用group by：默认select 参数为GroupBy的参数值, 可以用Cols或AllCols指定")
	users = []User{}
	Echor(CheckErr(engine.GroupBy("user_name").Cols("age", "user_name").Find(&users)), users)

	Echo("用having查询：")
	users = []User{}
	Echor(CheckErr(engine.Having("age=17").Find(&users)), users)

	Echo("Join连接多表查询：")
	gu := []GroupUser{}
	Echor(CheckErr(engine.Join("INNER", "group", "group.id=user.group_id").Find(&gu)), gu)

	Echo("回调方式的逐条查询：")
	CheckErr(engine.Where("age=17").Iterate(new(User), func(i int, bean interface{}) error {
		user := bean.(*User)
		Echor(i, *user)
		return nil
	}))

	Echo("迭代器的逐条查询：")
	user = new(User)
	rows := CheckErr(engine.Where("age=17").Rows(user)).(*xorm.Rows)
	defer rows.Close()
	for rows.Next() {
		Echor(CheckErr(rows.Scan(user)), user)
	}

	Echo("统计数据：")
	user = new(User)
	Echor(CheckErr(engine.Where("gender='女'").Count(user)))

	Echo("Sum系列方法：")
	Echo("Sum返回float64")
	Echor(CheckErr(engine.Sum(user, "age")))
	Echo("SumInt返回int64")
	Echor(CheckErr(engine.SumInt(user, "age")))
	Echo("Sums求多个字段的和，返回float64的Slice")
	Echor(CheckErr(engine.Sums(user, "age", "id")))
	Echo("Sums求多个字段的和，返回int64的Slice")
	Echor(CheckErr(engine.SumsInt(user, "age", "id")))

	Echo("删除：因为设置了deleted，所以不是真删除")
	Echor(CheckErr(engine.Delete(user)), *user)
}
