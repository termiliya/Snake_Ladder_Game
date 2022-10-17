package main

import (
	"Snake_Ladder_Game/server/slgmgo"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const MaxCircle = 10

type Game struct {
	RandNum int   `json:"randNum"`
	NowPos  int   `json:"nowPos"`
	Grid    []int `json:"grid"`
	Flag    bool  `json:"flag"`
}

func main() {
	slgmgo.MongoSetUp()
	e := gin.Default()
	e.GET("/dice/random", HandleDiceRandom)
	e.GET("/grid/init", HandleGridInit)
	e.GET("/slg/replay", HandleSLGReply)
	err := e.Run(":8861")
	if err != nil {
		log.Fatal(err)
	}
}

func HandleDiceRandom(c *gin.Context) {
	flag, move := DoClickDice()
	g := Game{NowPos: coordinate, RandNum: move, Flag: flag, Grid: grid}
	c.JSON(http.StatusOK, g)
}

func HandleGridInit(c *gin.Context) {
	grid = []int{
		30, -1, -1, -1, -1, 7,
		-1, -1, -1, -1, -1, -1,
		-1, -1, -1, 28, -1, -1,
		-1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1,
	}
	coordinate = 0
	g := Game{NowPos: coordinate, RandNum: 0, Flag: false, Grid: grid}

	incrId, err := slgmgo.InsertReply(grid, []int{})
	if err != nil {
		return
	}
	usingId = incrId
	c.JSON(http.StatusOK, g)
}

func HandleSLGReply(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		return
	}
	res, err := slgmgo.FindOneReply(id)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, res)
}

var (
	usingId    = 0
	coordinate = 0
	grid       = []int{
		-1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1,
	}
)

func DoClickDice() (bool, int) {
	rand.Seed(time.Now().Unix())
	move := rand.Intn(6-1) + 1
	var result, isUpdate bool
	result, coordinate, isUpdate = ChangeDiceRandomMap(move, coordinate, grid)
	fmt.Println(result, coordinate)
	if isUpdate {
		err := slgmgo.FindAndUpdateReply(usingId, move)
		if err != nil {
			log.Fatal(err)
		}
	}
	return result, move
}

// 是否到达终点，新的位置,是否更新骰子
func ChangeDiceRandomMap(move, coordinate int, grid []int) (bool, int, bool) {
	// 已经在终点了
	if coordinate == len(grid)-1 {
		return true, coordinate, false
	}
	// 新的位置
	newCoordinate := (move + coordinate) % len(grid)
	// 到终点
	if newCoordinate == len(grid)-1 {
		return true, newCoordinate, true
	}
	// 异常循环计数器
	errCount := 0
	// 有蛇/梯需再次移动
	for grid[newCoordinate] >= 0 {
		newCoordinate = grid[newCoordinate]
		// 超出上限，错误
		if newCoordinate > len(grid) {
			return true, newCoordinate, true
		}
		// 到终点
		if newCoordinate == len(grid)-1 {
			return true, newCoordinate, true
		}
		// 判断为到终点并抛错
		errCount++
		if errCount > MaxCircle {
			return true, newCoordinate, true
		}
	}
	return false, newCoordinate, true
}
