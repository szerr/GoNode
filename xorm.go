package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"log"
	"time"
)

var engine *xorm.Engine

type User struct { //如果和关键字冲突，需要加上单引号
	Id         int64
	Name       string `xorm:"varchar(32) index notnull 'user_name'"`
	Age        int
	Gender     string            `xorm:"varchar(8) default('男')"` //这里的中文 需要加个单引号
	Other      map[string]string `xorm:"json"`                    //复合数据类型可以用json存储
	CreateTime time.Time         `xorm:"created"`                 //创建时自动赋值
	DeleteTime time.Time         `xorm:"deleted"`                 //删除时自动赋值 and 不删除
	UpdateTime time.Time         `xorm:"updated"`                 //更新时自动赋值
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
	engine, err = xorm.NewEngine("mysql", "root@/xorm")
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
	print("\033[32m执行结果：")
	fmt.Println(args...)
	print("\033[0m")
}

func InsertTest(engine *xorm.Engine) int64 { //插入数据的例子
	var all_insert_num int64
	other := make(map[string]string)
	other["bula"] = "didi"
	other["didi"] = "bula"
	user1 := User{
		Name:   "bulabula",
		Age:    17,
		Gender: "男",
		Other:  other,
	}
	user2 := User{
		Name:   "didadida",
		Age:    18,
		Gender: "女",
		Other:  other,
	}
	insert_num, err := engine.Insert(&user1, &user2) //插两条
	CheckErr(err)
	all_insert_num += insert_num
	user2.Id += 1
	user2.Gender = "男"
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

	user := new(User)
	users := []User{}

	Echo("先清掉所有数据：")
	Echor(CheckErr(engine.Query("delete from user")))

	//插入数据：
	Echo("插入测试")
	Echo("插入数量: ", InsertTest(engine))

	//查询数据：
	Echo("查询第一条：")
	Echor(CheckErr(engine.Get(user)), *user)
	Echo("查询所有：")
	//Echor(CheckErr(engine.Find(&users)), users)
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

	user = new(User)
	Echo("用主键查询：")
	Echor(CheckErr(engine.Id(1).Get(user)), *user)

	//Echo("复合主键？这里没有这种东西")
	//engine.Id(core.PK{1, "name"}).Get(&user)
	//这里的两个参数按照struct中pk标记字段出现的顺序赋值。

	user = new(User)
	Echo("指定字段：")
	Echor(CheckErr(engine.Select("user_name, age").Id(1).Get(user)), *user)

	Echo("直接跑sql, 这里可以不清空struct")
	Echor(CheckErr(engine.Sql("select * from user where id=2").Get(user)), *user)

	Echo("某字段在一些值中：")
	Echor(CheckErr(engine.In("id", 1, 2, 3).Find(&users)), users)

	user = new(User)
	Echo("用Cols指定查询某些字段：")
	Echor(CheckErr(engine.Cols("age", "user_name").Get(user)), *user)
	Echo("用Cols指定更新某些字段：")
	user = new(User)
	user.Name = "haha"
	Echor(CheckErr(engine.Where("id=?", 1).Cols("user_name", "age").Update(user)), *user)

	Echo("指定操作所有字段：一般与Update配合使用，因为默认Update只更新非0，非”“，非bool的字段。这里只做个查询 = = ")
	Echor(CheckErr(engine.Where("id=?", 2).AllCols().Update(user)), *user)

	Echo("必须更新某些字段：让Update更新非0，非”“，非bool的字段。这个功能和Cols一样")
	Echor(CheckErr(engine.Where("id=?", 3).MustCols("age").Update(user)), *user)

	//再插一坨数据
	InsertTest(engine)

	Echo("删掉所有数据：")
	Echor(CheckErr(engine.Delete(user)), user)
}
