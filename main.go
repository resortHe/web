package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	_ "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
)

type Test struct {
	Id              int    `json:"id" gorm:"primaryKey"`
	NameEncrypted   string `json:"name_encrypted"`
	NameUnencrypted string `json:"name_unencrypted"`
	AgeEncrypted    string `json:"age_encrypted"`
	AgeUnencrypted  string `json:"age_unencrypted"`
}

var db *gorm.DB

func init() {
	var err error
	dsn := "root:123456@tcp(192.168.6.197:9399)/test?parseTime=true"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(&Test{})
	if err != nil {
		log.Fatal(err)
	}
}
func main() {
	r := gin.Default()
	r.Use(cors.Default())
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.GET("/list", func(c *gin.Context) {
		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}
		pageSize := 20
		offset := (page - 1) * pageSize
		test := make([]Test, pageSize)
		var totalCount int64
		if err := db.Table("test").Count(&totalCount).Error; err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from the database"})
			return
		}

		if err := db.Table("test").Offset(offset).Limit(pageSize).Find(&test).Error; err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from the database"})
			return
		}

		log.Printf("Retrieved tests: %+v\n", test)
		hasNextPage := int64(page*pageSize) < totalCount

		c.JSON(http.StatusOK, gin.H{"tests": test, "hasNextPage": hasNextPage})
	})

	r.POST("/add", func(c *gin.Context) {
		var test Test
		err := c.ShouldBindJSON(&test)
		if err != nil {
			fmt.Printf("Error binding JSON: %s\n", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
			return
		}
		fmt.Printf("Received test data: %+v\n", test)

		err = db.Table("test").Create(&test).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user to the database"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "User added successfully"})
	})

	err := r.Run("192.168.6.163:8080")
	if err != nil {
		fmt.Println(err)
	}
}
