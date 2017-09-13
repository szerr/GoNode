package main

import (
	"github.com/go-bongo/bongo"
	"gopkg.in/mgo.v2/bson"
	"log"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
}

type Results struct {
	Math    int
	English int
	Langu   int
	Sports  int
}

type Student struct { //设定文档的数据结构
	bongo.DocumentBase `bson:",inline"`
	Name               string
	Age                int
	Score              Results
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
	log.Println(args...)
	print("\033[0m")
}

func CheckErr(args ...interface{}) interface{} {
	err, _ := args[len(args)-1].(error)
	if err != nil {
		log.Fatal(err)
	}
	return args[0]
}

func InsertData(cc *bongo.Collection) []*Student {
	li := []*Student{}
	for _, i := range []int{12, 10, 18, 17} {
		te := &Student{
			Name: "秀吉",
			Age:  i,
			Score: Results{
				Math:    12,
				English: 10,
				Langu:   11,
				Sports:  150,
			},
		}
		CheckErr(cc.Save(te))
		li = append(li, te)
	}
	return li
}

func main() {
	conn, err := bongo.Connect(&bongo.Config{
		ConnectionString: "", //默认是本机
		Database:         "bongo",
	})
	CheckErr(err)

	cc := conn.Collection("student")
	cc.Delete(bson.M{}) //清掉所有数据
	Echo("在student存一份学生数据")
	li := InsertData(cc)
	Echo("根据ID查询")
	student := new(Student)
	id := li[0].Id.Hex() //获取id的字符串
	Echor(CheckErr(cc.FindById(bson.ObjectIdHex(id), student)), *student)
	Echo("查找多条：因为数据太少了，这里就查所有吧")
	results := cc.Find(bson.M{})
	for results.Next(student) {
		Echor(student)
	}
	Echo("根据属性查找一条")
	Echor(CheckErr(cc.FindOne(bson.M{"age": 10}, student)), *student)
	Echo("把10岁的秀吉边变16岁")
	student.Age = 16
	cc.Save(student)
	Echo("删掉文档：")
	Echor(CheckErr(cc.DeleteDocument(student)))
	Echo("删掉18岁的秀吉：这里返回mgo.ChangeInfo 存有更新 移除 匹配的数量")
	Echor(CheckErr(cc.Delete(bson.M{"age": 18})))
}
