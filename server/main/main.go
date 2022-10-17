package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

//func main(){
//	//mp := map[int]int{1:1,2:2}
//	sl :=[]int{1,2,3}
//	for k,v:= range sl{
//		go func(){
//			fmt.Println(k,v)
//			//f(k,v)
//		}()
//		//fmt.Println(k,v)
//		//go func(k,v int){
//		//	fmt.Println(k,v)
//		//	f(k,v)
//		//}(k,v)
//	}
//	time.Sleep(time.Second)
//}
//
//func f(k,v int){fmt.Println(k,v)}

const MaxCircle = 10

type Game struct {
	RandNum int   `json:"randNum"`
	NowPos  int   `json:"nowPos"`
	Grid    []int `json:"grid"`
	Flag    bool  `json:"flag"`
}

func main() {
	http.HandleFunc("/dice/random", HandleDiceRandom)
	http.HandleFunc("/grid/init", HandleGridInit)
	err := http.ListenAndServe(":8861", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func HandleDiceRandom(w http.ResponseWriter, r *http.Request) {
	flag, move := DoClickDice()
	g := Game{NowPos: coordinate, RandNum: move, Flag: flag, Grid: grid}
	bytes, err := json.Marshal(g)
	if err != nil {
		return
	}
	_, err = io.WriteString(w, string(bytes))
	if err != nil {
		return
	}
}

func HandleGridInit(w http.ResponseWriter, r *http.Request) {
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
	bytes, err := json.Marshal(g)
	if err != nil {
		return
	}
	_, err = io.WriteString(w, string(bytes))
	if err != nil {
		return
	}
}

var (
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
	var result bool
	result, coordinate = ChangeDiceRandomMap(move, coordinate, grid)
	fmt.Println(result, coordinate)
	return result, move
}

// 是否到达终点，新的位置
func ChangeDiceRandomMap(move, coordinate int, grid []int) (bool, int) {
	// 已经在终点了
	if coordinate == len(grid)-1 {
		return true, coordinate
	}
	// 新的位置
	newCoordinate := (move + coordinate) % len(grid)
	// 到终点
	if newCoordinate == len(grid)-1 {
		return true, newCoordinate
	}
	// 异常循环计数器
	errCount := 0
	// 有蛇/梯需再次移动
	for grid[newCoordinate] >= 0 {
		newCoordinate = grid[newCoordinate]
		// 超出上限，错误
		if newCoordinate > len(grid) {
			return true, newCoordinate
		}
		// 到终点
		if newCoordinate == len(grid)-1 {
			return true, newCoordinate
		}
		// 判断为到终点并抛错
		errCount++
		if errCount > MaxCircle {
			return true, newCoordinate
		}
	}
	return false, newCoordinate
}
