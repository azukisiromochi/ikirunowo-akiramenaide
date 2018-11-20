package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"

	_ "github.com/azukisiromochi/ikirunowo-akiramenaide/statik"
	"github.com/kyokomi/lottery"
	"github.com/rakyll/statik/fs"
	"golang.org/x/text/width"
)

// MixedPachi は１種２種混合機を意識した遊技台
type MixedPachi struct {
	// 確率分子（電サポ時）
	Numer int
	// 確率分母（電サポ時）
	Denom int

	// RUSH 抽選用モード時の規定ゲーム数（保留も含む）
	ChallengeLimit int
	// RUSH 中の規定ゲーム数（保留も含む）
	RushLimit int

	// RUSH 継続数
	RushCount int
	// 総獲得 pt 数
	Points int32
}

// 大当りの種類（16R, 9R, 4R, みたいなの）
type feverKind struct {
	// 獲得 pt （期待値）
	Point int32
	// 図柄（７とか１とか）
	DispDesign string
	// 当選割合（’16R は 40%’とかの "40" 部分を設定）
	DropProb int
}

// lottery 使用のために必要な抽選用のメソッド
func (f feverKind) Prob() int {
	return f.DropProb
}

// NewSymphogear は『CRF戦姫絶唱シンフォギア』用のインスタンス生成関数
func NewSymphogear() *MixedPachi {
	return &MixedPachi{
		// 1/7.4 に近しい値を設定
		Numer: 10,
		Denom: 74,
		// 電サポ１回 + 保留玉4個
		ChallengeLimit: 5,
		// 電サポ１回 + 保留玉4個
		RushLimit: 11,
		// 通常時の大当たり数からスタート
		RushCount: 1,
		// 通常時の大当たりの 99% が 4R のため, その期待値を設定
		Points: 392,
	}
}

// Challenge は RUSH 抽選用モードをスタートする
// 当選時は true を返却するので, それをトリガーに Rush メソッドを呼び出すこと.
// 非当選時は false を返す.
func (mp *MixedPachi) Challenge() bool {

	msg("/challenge-start.txt")

	stdin := bufio.NewScanner(os.Stdin)
	for i := 0; i < mp.ChallengeLimit; i++ {

		msg("/challenge-draw.txt")
		stdin.Scan()

		if mp.draw() {
			mp.Points += dispFever()
			mp.RushCount++
			return true
		} else {
			dispNoFever()
		}
	}

	// 非当選時は Failed を流して終了
	msg("/challenge-failed.txt")
	fmt.Println()
	return false
}

// Rush をスタートする.
// 引数に AA 用のファイルパスを持ち, 開始か継続かを指定する.
// 当選時は true を, 非当選時は false を返す.
// MEMO:
// Rush の仕様は遊技台に大きく依存するため, 現状は『CRF戦姫絶唱シンフォギア』の
// 仕様に準ずる使用になっている.
func (mp *MixedPachi) Rush(filePath string) bool {

	msg(filePath)

	stdin := bufio.NewScanner(os.Stdin)
	for i := 0; i < mp.RushLimit-5; i++ {

		stdin.Scan()

		if mp.draw() {
			mp.Points += dispFever()
			mp.RushCount++
			return true
		} else {
			dispNoFever()
		}
	}

	// 表示上の最終ゲーム抽選（保留を含む4ゲーム分の抽選）
	msg("/rush-last-draw.txt")
	stdin.Scan()

	last := false
	for i := 0; i < 4; i++ {

		// 当選 / 非当選時の図柄表示はゲーム数をまとめている都合上表示しない.
		// ループ終了後に図柄表示
		if mp.draw() {
			last = true
			break
		}
	}
	if last {
		mp.Points += dispFever()
		mp.RushCount++
		return true
	} else {
		dispNoFever()
	}

	// システム上の最終ゲーム抽選（まだ響と流れ星見てない！）
	if mp.draw() {
		time.Sleep(3 * time.Second)
		msg("/rush-shooting-star.txt")
		time.Sleep(1 * time.Second)
		mp.Points += dispFever()
		mp.RushCount++
		return true
	}

	// 非当選
	return false
}

// End は Rush 終了時の表示を行う.
// Rush 継続数と総獲得枚数の集計, 表示がメイン.
func (mp *MixedPachi) End() {
	time.Sleep(500 * time.Millisecond)
	msg("/rush-end.txt")

	// Rush 継続数を表示用に編集
	rushCount := width.Widen.String(strconv.Itoa(mp.RushCount))
	if utf8.RuneCountInString(rushCount) == 1 {
		rushCount = rushCount + "　　"
	} else if utf8.RuneCountInString(rushCount) == 2 {
		rushCount = rushCount + "　"
	}

	// 総獲得枚数を表示用に編集
	totalPoints := width.Widen.String(fmt.Sprint(mp.Points))
	preSpace := ""
	if utf8.RuneCountInString(totalPoints) == 3 {
		preSpace = "　　　   " + preSpace
	} else if utf8.RuneCountInString(totalPoints) == 4 {
		preSpace = "　　  " + preSpace
	} else if utf8.RuneCountInString(totalPoints) == 5 {
		preSpace = "　 " + preSpace
	}
	// 各数字の間に半角スペース埋め & 前全角スペース埋め
	re, _ := regexp.Compile(`(.)`)
	dispTotalPoints := preSpace + re.ReplaceAllString(totalPoints, "${1} ")

	fmt.Println("　　　　　　　　　　┌ーーーーーーーーーーーーーーーーーー┐")
	fmt.Println("　　　　　　　　　　│　　　　　　　　　　　　　　　　　　│")
	fmt.Printf("　　　　　　　　　　│　　Ｆ Ｅ Ｖ Ｅ Ｒ　✕　%s　　　│\n", rushCount)
	fmt.Println("　　　　　　　　　　│　　　　　　　　　　　　　　　　　　│")
	fmt.Println("　　　　　　　　　　│　　Ｔ Ｏ Ｔ Ａ Ｌ　　　　　　　　　│")
	fmt.Printf("　　　　　　　　　　│　　　　 %s ｐ ｔ 　│\n", dispTotalPoints)
	fmt.Println("　　　　　　　　　　│　　　　　　　　　　　　　　　　　　│")
	fmt.Println("　　　　　　　　　　└ーーーーーーーーーーーーーーーーーー┘")
	fmt.Println()
}

// 抽選処理.
// 画面への表示は行わなく単純な抽選のみ.
// 当選は true, 非当選は false.
func (mp *MixedPachi) draw() bool {
	lot := lottery.NewDefault()
	return lot.LotOf(mp.Numer, mp.Denom)
}

// 当選時の図柄を表示する.
// 電サポ中の当選はラウンド数が次のように振分けられる.
//   実質15R：約1470個（40%）SPECIAL FEVER
//   実質12R：約1176個（3%）ギアV-COMBO
//   実質8R：約784個（7%）ギアV-COMBO
//   実質4R：約392個（50%）FEVER
// 図柄表示とともに期待出玉数を返す.
func dispFever() int32 {

	ferverkinds := []lottery.Interface{
		feverKind{Point: 1470, DispDesign: "７", DropProb: 40},
		feverKind{Point: 1176, DispDesign: "３", DropProb: 3},
		feverKind{Point: 784, DispDesign: "５", DropProb: 7},
		feverKind{Point: 392, DispDesign: "１", DropProb: 10},
		feverKind{Point: 392, DispDesign: "２", DropProb: 10},
		feverKind{Point: 392, DispDesign: "３", DropProb: 5},
		feverKind{Point: 392, DispDesign: "４", DropProb: 10},
		feverKind{Point: 392, DispDesign: "５", DropProb: 5},
		feverKind{Point: 392, DispDesign: "６", DropProb: 10},
	}

	// ラウンド数抽選
	lot := lottery.NewDefault()
	lotIdx := lot.Lots(ferverkinds...)
	ferver := ferverkinds[lotIdx].(feverKind)

	// 図柄表示
	time.Sleep(50 * time.Millisecond)
	fmt.Print(ferver.DispDesign, " ")
	time.Sleep(50 * time.Millisecond)
	fmt.Print(ferver.DispDesign, " ")
	time.Sleep(100 * time.Millisecond)
	fmt.Println(ferver.DispDesign)

	// 期待出玉数返却
	return ferver.Point
}

// 非当選時の図柄を表示する.
// 誤って当選しないように制御を加えるとともに,
// 左に"７"が止まると「激アツ」演出になってしまうため制御する.
func dispNoFever() {
	ignore := ""
	// 左
	d1 := drawDesign("７")
	time.Sleep(50 * time.Millisecond)
	fmt.Print(d1, " ")
	// 中（ホントは右がいいけど...）
	d2 := drawDesign(ignore)
	time.Sleep(50 * time.Millisecond)
	fmt.Print(d2, " ")
	// 右
	if d1 == d2 {
		ignore = d1
	}
	time.Sleep(100 * time.Millisecond)
	d3 := drawDesign(ignore)
	fmt.Println(d3)
}

// 非当選時の図柄を決定する.
// 引数の図柄を除くものから抽選する.
func drawDesign(ignore string) string {
	rand.Seed(time.Now().UnixNano())
	var num string
	for {
		// 半角 => 全角
		num = width.Widen.String(strconv.Itoa(rand.Intn(7) + 1))
		if num != ignore {
			return num
		}
	}
}

// AA 表示用の出力.
// 指数のファイルパスを読み込み, 一行ずつ出力する.
// statik の利用に伴い, ファイルパスは "/fileName.txt" のように指定.
func msg(filePath string) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	file, err := statikFS.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
	}
}
