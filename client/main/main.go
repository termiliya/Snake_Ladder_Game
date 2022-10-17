package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type Game struct {
	RandNum int   `json:"randNum"`
	NowPos  int   `json:"nowPos"`
	Grid    []int `json:"grid"`
	Flag    bool  `json:"flag"`
	mode    Mode
}

type Mode int

const (
	ModeTitle Mode = iota
	ModeGame
	ModeGameOver
)

const (
	LineNum     = 6
	RowNum      = 6
	GridWidth   = 360
	GridHeight  = 360
	BlockWidth  = GridWidth / LineNum
	BlockHeight = GridHeight / RowNum
)
const (
	NumFontSize   = 20
	TitleFontSize = 20
)

var (
	mplusNumFont   font.Face
	mplusTitleFont font.Face
)

func (g *Game) Get(url string) {
	// 请求后端接口
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:8861%s", url))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var game Game
	err = json.Unmarshal(bytes, &game)
	if err != nil {
		log.Fatal(err)
	}
	g.RandNum = game.RandNum
	g.NowPos = game.NowPos
	g.Grid = game.Grid
	g.Flag = game.Flag
}

func (g *Game) Update() error {
	switch g.mode {
	case ModeTitle:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.Get("/grid/init")
			g.mode = ModeGame
		}
	case ModeGame:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			// 请求后端接口
			g.Get("/dice/random")
			if g.Flag {
				g.mode = ModeGameOver
			}
		}
	case ModeGameOver:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.mode = ModeTitle
		}
	}
	return nil
}

func (g *Game) DrawModeTitle(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 0xbb, G: 0xad, B: 0xa0, A: 0xff})
	text.Draw(screen, "PRESS SPACE KEY START", mplusTitleFont, 60, 5*TitleFontSize, color.White)
}

func (g *Game) DrawModeGame(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 0xbb, G: 0xad, B: 0xa0, A: 0xff})
	f := mplusNumFont
	// 画网格
	for i := 0; i < LineNum; i++ {
		ebitenutil.DrawRect(screen, float64(BlockWidth*i), 0, 10, GridHeight, color.RGBA{R: 0x80, G: 0xa0, B: 0xc0, A: 0xff})
	}
	for i := 0; i < RowNum; i++ {
		ebitenutil.DrawRect(screen, 0, float64(BlockHeight*i), GridWidth, 10, color.RGBA{R: 0x80, G: 0xa0, B: 0xc0, A: 0xff})
	}
	// 画数字
	for row := 0; row < RowNum; row++ {
		for line := 0; line < LineNum; line++ {
			pos := row*LineNum + line
			y := (RowNum - 1 - row) * BlockHeight
			x := line * BlockWidth
			if row%2 == 1 {
				x = (LineNum - 1 - line) * BlockWidth
			}
			text.Draw(screen, strconv.Itoa(pos+1), f, x+10, y+30, tileColor(2))
			if g.NowPos == pos {
				text.Draw(screen, strconv.Itoa(pos+1), f, x+10, y+30, tileColor(8))
			}
			// 画梯子
			lePos := g.Grid[pos]
			if lePos >= 0 {
				tmpRow := lePos / LineNum
				tmpLine := lePos % LineNum
				if tmpRow%2 == 1 {
					tmpLine = LineNum - 1 - (lePos % LineNum)
				}
				ley := (RowNum - 1 - tmpRow) * BlockHeight
				leX := tmpLine * BlockWidth
				// 梯子
				f, err := os.Open("./client/img/ladder.png")
				if err != nil {
					log.Fatal(err)
				}
				img, _, err := image.Decode(f)
				if err != nil {
					log.Fatal(err)
				}
				lI := ebiten.NewImageFromImage(img)
				w, h := lI.Size()
				op := &ebiten.DrawImageOptions{}
				// 角度，长度，偏移
				baseRate := 1.0 // 60 * 60
				newRate := math.Sqrt(math.Pow(float64(ley-y), 2)+math.Pow(float64(leX-x), 2)) / BlockHeight * baseRate
				// 图片长度
				op.GeoM.Scale(baseRate, newRate)
				// 位移图片中心方便旋转
				op.GeoM.Translate(-float64(w)*baseRate/2, -float64(h)*newRate/2)
				// 旋转角度
				theta := math.Pi/2 - math.Atan(-float64(ley-y)/float64(leX-x))
				op.GeoM.Rotate(theta)
				//// 移动图片到新的位置
				op.GeoM.Translate(float64(leX+x)/2+30, float64(ley+y)/2+30)

				screen.DrawImage(lI, op)
			}
		}
	}
	//在屏幕上输出
	text.Draw(screen, fmt.Sprintf("RANDOM: %d, NOWNUM: %d", g.RandNum, g.NowPos+1), mplusTitleFont, 30, 400, color.White)
}

func (g *Game) DrawModeGameOver(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 0xbb, G: 0xad, B: 0xa0, A: 0xff})
	text.Draw(screen, "", mplusTitleFont, 60, 4*TitleFontSize, color.White)
	if g.Flag == true && g.NowPos == len(g.Grid)-1 {
		text.Draw(screen, "GAME SUCCESS", mplusTitleFont, 60, 5*TitleFontSize, color.White)
		text.Draw(screen, fmt.Sprintf("RANDOM: %d, NOWNUM: %d", g.RandNum, len(g.Grid)), mplusTitleFont, 30, 7*TitleFontSize, color.White)
	} else {
		text.Draw(screen, "GAME OVER", mplusTitleFont, 60, 5*TitleFontSize, color.White)
	}
	text.Draw(screen, "PRESS SPACE TO BACK", mplusTitleFont, 30, 6*TitleFontSize, color.White)
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.mode {
	case ModeTitle:
		g.DrawModeTitle(screen)
	case ModeGame:
		g.DrawModeGame(screen)
	case ModeGameOver:
		g.DrawModeGameOver(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 360, 400 //窗口分辨率
}

func newGame() *Game {
	return &Game{
		RandNum: 0,
		Grid: []int{
			0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0,
		},
	}
}

func tileColor(value int) color.Color {
	switch value {
	case 2, 4:
		return color.RGBA{R: 0x77, G: 0x6e, B: 0x65, A: 0xff}
	case 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536:
		return color.RGBA{R: 0xf9, G: 0xf6, B: 0xf2, A: 0xff}
	}
	panic("not reach")
}

func main() {
	// 文本
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}
	mplusNumFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    NumFontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	mplusTitleFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    TitleFontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	ebiten.SetWindowSize(GridWidth, GridHeight) //窗口大小
	ebiten.SetWindowTitle("Snake Ladder Game")  //窗口标题
	if err := ebiten.RunGame(newGame()); err != nil {
		log.Fatal(err)
	}
}
