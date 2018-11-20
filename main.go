package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()

	app.Action = func(c *cli.Context) error {

		stdin := bufio.NewScanner(os.Stdin)

		for {
			playSymphogear()
			for {
				time.Sleep(1 * time.Second)
				fmt.Println("もう一度遊戯しますか？ (Y/n)")
				stdin.Scan()
				isNextPlay := stdin.Text()
				if isNextPlay == "y" || isNextPlay == "Y" || isNextPlay == "ｙ" || isNextPlay == "Ｙ" {
					break
				}
				if isNextPlay == "n" || isNextPlay == "N" || isNextPlay == "ｎ" || isNextPlay == "Ｎ" {
					return nil
				}
			}
		}
	}

	app.Run(os.Args)
}

// 『CRF戦姫絶唱シンフォギア』を遊戯する
func playSymphogear() {

	// ？？？「シンフォギアァァァァァァアッ！！！」
	symphogear := NewSymphogear()

	// ？？？「逆さ鱗に触れたのだ、相応の覚悟はできておろうな」
	if symphogear.Challenge() {

		// シンフォギアチャンス突入！
		if symphogear.Rush("/rush-start.txt") {
			for {
				// シンフォギアチャンス継続！！
				if !symphogear.Rush("/rush-continue.txt") {
					break
				}
			}
		}
		// お疲れ様でした
		symphogear.End()
	}
}
