package main

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func initDb() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/dafanji?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("failed to connect database")
	}
	sqlDb, err := DB.DB()
	if err != nil {
		fmt.Println("failed to connect database")
	}
	sqlDb.SetMaxIdleConns(10)           // 设置连接池，默认值为10
	sqlDb.SetMaxOpenConns(100)          // 设置数据库的最大打开连接数，默认值为100
	sqlDb.SetConnMaxLifetime(time.Hour) // 设置连接可重用的最大时间，默认值为0
}

//假设有一个名为 students 的表，包含字段 id （主键，自增）、 name （学生姓名，字符串类型）、 age （学生年龄，整数类型）、 grade （学生年级，字符串类型）。
//要求 ：
//编写SQL语句向 students 表中插入一条新记录，学生姓名为 "张三"，年龄为 20，年级为 "三年级"。
//编写SQL语句查询 students 表中所有年龄大于 18 岁的学生信息。
//编写SQL语句将 students 表中姓名为 "张三" 的学生年级更新为 "四年级"。
//编写SQL语句删除 students 表中年龄小于 15 岁的学生记录。

type Student struct {
	Id    int
	Name  string
	Age   int
	Grade string
}

func One() {

	DB.AutoMigrate(&Student{})

	// 1. 插入记录
	student := Student{Name: "张三", Age: 20, Grade: "三年级"}
	result := DB.Create(&student)
	fmt.Printf("插入记录ID: %d, 影响行数: %d, 错误: %v\n", student.Id, result.RowsAffected, result.Error)

	// 2. 查询年龄大于18的学生
	var students []Student
	DB.Where("age > ?", 18).Find(&students)
	fmt.Printf("年龄大于18的学生: %+v\n", students)

	// 3. 更新张三年级为四年级
	updateResult := DB.Model(&Student{}).Where("name = ?", "张三").Update("grade", "四年级")
	fmt.Printf("更新影响行数: %d, 错误: %v\n", updateResult.RowsAffected, updateResult.Error)

	// 4. 删除年龄小于15岁的学生
	deleteResult := DB.Where("age < ?", 15).Delete(&Student{})
	fmt.Printf("删除影响行数: %d, 错误: %v\n", deleteResult.RowsAffected, deleteResult.Error)

}

// 假设有两个表： accounts 表（包含字段 id 主键， balance 账户余额）和 transactions 表（包含字段 id 主键， from_account_id 转出账户ID， to_account_id 转入账户ID， amount 转账金额）。
// 要求 ：
// 编写一个事务，实现从账户 A 向账户 B 转账 100 元的操作。在事务中，需要先检查账户 A 的余额是否足够，如果足够则从账户 A 扣除 100 元，向账户 B 增加 100 元，并在 transactions 表中记录该笔转账信息。如果余额不足，则回滚事务。
type Account struct {
	Id      int
	Balance float64
}

type Transaction struct {
	Id            int
	FromAccountId int
	ToAccountId   int
	Amount        float64
}

func Two() {
	err := DB.AutoMigrate(&Account{}, &Transaction{})
	if err != nil {
		return
	}
	err1 := DB.Transaction(func(tx *gorm.DB) error {
		var accountA, accountB Account
		DB.Create(&Account{Id: 1, Balance: 1000})
		DB.Create(&Account{Id: 2, Balance: 1000})
		DB.Where("id = ?", 1).First(&accountA)
		DB.Where("id = ?", 2).First(&accountB)
		if accountA.Balance < 100 {
			return fmt.Errorf("账户余额不足")
		}
		accountA.Balance -= 100
		accountB.Balance += 100
		DB.Save(&accountA)
		DB.Save(&accountB)
		return nil
	})
	if err1 != nil {
		return
	}

}

//假设你已经使用Sqlx连接到一个数据库，并且有一个 employees 表，包含字段 id 、 name 、 department 、 salary 。
//要求 ：
//编写Go代码，使用Sqlx查询 employees 表中所有部门为 "技术部" 的员工信息，并将结果映射到一个自定义的 Employee 结构体切片中。
//编写Go代码，使用Sqlx查询 employees 表中工资最高的员工信息，并将结果映射到一个 Employee 结构体中。

type Employee struct {
	Id         int
	Name       string
	Department string
	Salary     float64
}

func Three() {
	//DB.AutoMigrate(&Employee{})

	var employees []Employee
	DB.Where("department = ?", "技术部").Find(&employees)
	fmt.Printf("技术部员工信息: %+v\n", employees)

	var employee Employee
	DB.Order("salary DESC").Limit(1).Find(&employee)
	fmt.Printf("工资最高的员工信息: %+v\n", employee)

	//技术部员工信息: [{Id:1 Name:张三 Department:技术部 Salary:1000} {Id:3 Name:王五 Department:技术部 Salary:3000}]
	//工资最高的员工信息: {Id:3 Name:王五 Department:技术部 Salary:3000}
}

//假设有一个 books 表，包含字段 id 、 title 、 author 、 price 。
//要求 ：
//定义一个 Book 结构体，包含与 books 表对应的字段。
//编写Go代码，使用Sqlx执行一个复杂的查询，例如查询价格大于 50 元的书籍，并将结果映射到 Book 结构体切片中，确保类型安全。

type Book struct {
	Id     int
	Title  string
	Author string
	Price  float64
}

func Four() {
	//DB.AutoMigrate(&Book{})
	//DB.Create(&Book{Id: 1, Title: "《Go 语言基础》", Author: "小王", Price: 80})
	//DB.Create(&Book{Id: 2, Title: "《Go 语言进阶》", Author: "小子", Price: 100})
	//DB.Create(&Book{Id: 3, Title: "《Go 语言实战》", Author: "王子", Price: 300})
	//DB.Create(&Book{Id: 4, Title: "《Go 语言微服务》", Author: "小王子", Price: 500})
	var books []Book
	DB.Where("price > ?", 50).Find(&books)
	fmt.Printf("价格大于50元的书籍: %+v\n", books)
}

// 假设你要开发一个博客系统，有以下几个实体： User （用户）、 Post （文章）、 Comment （评论）。
// 要求 ：
// 使用Gorm定义 User 、 Post 和 Comment 模型，其中 User 与 Post 是一对多关系（一个用户可以发布多篇文章）， Post 与 Comment 也是一对多关系（一篇文章可以有多个评论）。
// 编写Go代码，使用Gorm创建这些模型对应的数据库表。

type User struct {
	Id        int
	Name      string
	Posts     []Post
	PostCount int `gorm:"default:0"`
}

type Post struct {
	Id            int
	Title         string
	Content       string
	Comments      []Comment
	UserId        int
	CommentStatus string `gorm:"default:'有评论'"`
}

type Comment struct {
	Id      int
	Content string
	PostId  int
}

func Five() {
	//DB.AutoMigrate(&User{}, &Post{}, &Comment{})
	DB.Create(&User{Id: 1, Name: "小王1"})
	DB.Create(&Post{Id: 1, Title: "《Go 语言基础》", Content: "Go 语言基础", UserId: 1})
	DB.Create(&Comment{Id: 1, Content: "Go 语言基础不错", PostId: 1})
	DB.Create(&User{Id: 2, Name: "小王2"})
	DB.Create(&Post{Id: 2, Title: "《Go 语言进阶》", Content: "Go 语言进阶", UserId: 1})
	DB.Create(&Comment{Id: 2, Content: "Go 语言进阶不错", PostId: 2})
	DB.Create(&User{Id: 3, Name: "小王3"})
	DB.Create(&Post{Id: 3, Title: "《Go 语言实战》", Content: "Go 语言实战", UserId: 1})
	DB.Create(&Comment{Id: 3, Content: "Go 语言实战不错", PostId: 3})
	DB.Create(&User{Id: 4, Name: "小王"})
	DB.Create(&Post{Id: 4, Title: "《Go 语言微服务》", Content: "Go 语言微服务", UserId: 2})
	DB.Create(&Comment{Id: 4, Content: "Go 语言微服务不错", PostId: 4})
	DB.Create(&User{Id: 5, Name: "小王"})
	DB.Create(&Post{Id: 5, Title: "《Go 语言微服务》", Content: "Go 语言微服务", UserId: 3})
	DB.Create(&Comment{Id: 5, Content: "Go 语言微服务不错", PostId: 4})
	DB.Create(&User{Id: 6, Name: "小王"})
	DB.Create(&Post{Id: 6, Title: "《Go 语言微服务》", Content: "Go 语言微服务", UserId: 4})
	DB.Create(&Comment{Id: 6, Content: "Go 语言微服务不错", PostId: 4})
}

//基于上述博客系统的模型定义。
//要求 ：
//编写Go代码，使用Gorm查询某个用户发布的所有文章及其对应的评论信息。
//编写Go代码，使用Gorm查询评论数量最多的文章信息。

func Six() {
	var user User
	DB.Preload("Posts").Preload("Posts.Comments").Where("id = ?", 1).Find(&user)

	fmt.Printf("用户%s的博客文章和评论信息: %+v\n", user.Name, user.Posts)
	//编写Go代码，使用Gorm查询评论数量最多的文章信息。
	var post Post
	DB.Preload("Comments").Model(&Post{}).Select("posts.*, count(comments.id) as comments_count").Joins("left join comments on posts.id = comments.post_id").Group("posts.id").Order("comments_count DESC").Limit(1).Find(&post)
	fmt.Printf("评论数量最多的文章信息: %+v\n", post)
}

//以上面的内容为基础 ：
//以上面的内容为基础，为 Post 模型添加一个钩子函数，在文章创建时自动更新用户的文章数量统计字段。

// 为 Comment 模型添加一个钩子函数，在评论删除时检查文章的评论数量，如果评论数量为 0，则更新文章的评论状态为 "无评论"。
// 为 Post 模型添加创建钩子函数
func (p *Post) AfterCreate(tx *gorm.DB) error {
	// 在文章创建后，更新用户的文章数量统计
	err := tx.Model(&User{}).Where("id = ?", p.UserId).UpdateColumn("post_count", gorm.Expr("post_count + ?", 1)).Error
	return err
}

// 如果需要在文章删除时也更新用户的文章数量统计，可以添加删除钩子
func (p *Post) AfterDelete(tx *gorm.DB) error {
	// 在文章删除后，更新用户的文章数量统计
	err := tx.Model(&User{}).Where("id = ?", p.UserId).UpdateColumn("post_count", gorm.Expr("post_count - ?", 1)).Error
	return err
}

// 为 Comment 模型添加删除钩子函数
func (c *Comment) AfterDelete(tx *gorm.DB) error {
	// 统计该文章的评论数量
	var count int64

	tx.Debug().Model(&Comment{}).Where("post_id = ?", c.PostId).Count(&count)

	// 如果评论数量为 0，则更新文章的评论状态为 "无评论"
	if count == 0 {
		err := tx.Debug().Model(&Post{}).Where("id = ?", c.PostId).Update("comment_status", "无评论").Error
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	initDb()
	//One()
	//Two()
	//Three()
	//Four()

	//Five()
	//Six()
	// 建议改为以下写法之一来确保钩子触发
	var comment Comment
	DB.Where("id = ?", 1).First(&comment)
	DB.Delete(&comment)
	//不会触发钩子，必须实例模型
	//DB.Preload("Posts").Delete(&Comment{},1)
}
