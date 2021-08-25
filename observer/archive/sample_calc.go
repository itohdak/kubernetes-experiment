package main

import (
	"fmt"
	"html"
	"math/big"
	"net/http"
)

// ServerHTTPメソッド用の構造体型

type Server struct{}

// httpリクエストを受け取るメソッド

func (Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// フォームの入力値を取得
	left := r.FormValue("left")
	right := r.FormValue("right")
	op := r.FormValue("op") //演算子（ラジオボタンの値）

	// 文字列を整数に変換
	leftInt := &big.Int{}
	rightInt := &big.Int{}

	_, leftOK := leftInt.SetString(left, 10)
	_, rightOK := rightInt.SetString(right, 10)

	// 四則演算の処理
	// 変換エラーがなければ、演算子に従って計算
	var result string
	if leftOK && rightOK {
		resultInt := &big.Int{}
		// 演算子ごとに分岐
		switch op {
		case "add":
			resultInt.Add(leftInt, rightInt)
		case "sub":
			resultInt.Sub(leftInt, rightInt)
		case "multi":
			resultInt.Mul(leftInt, rightInt)
		case "div":
			resultInt.Div(leftInt, rightInt)
		}
		result = resultInt.String()
	}

	//HTMLの文字列
	h := `
            <html>
                <head>
                    <title>電卓アプリ</title>
                </head>
                <body>
                    <form action ="/" method="post">
                        左項目：<input type="text" name="left"><br>
                        右項目：<input type="text" name="right"><br>
                        演算子：
                        <input type="radio" name="op" value="add" checked> +
                        <input type="radio" name="op" value="sub" > -
                        <input type="radio" name="op" value="multi" > ×
                        <input type="radio" name="op" value="div" > ÷
                        <br><input type="submit" name="送信"><hr>

                        [フォームの入力値]<br>
                        左項目：` + html.EscapeString(left) + `<br>
                        右項目：` + html.EscapeString(right) + `<br>
                        演算子：` + html.EscapeString(op) + `<br>
                        演算結果：` + html.EscapeString(result) + `<br>
                    </form>
                </body>
            </html>
        `
	// クライアント（ブラウザ）にHTMLを送信
	fmt.Fprint(w, h)
}

func main() {
	// Webサーバを起動
	http.ListenAndServe(":4020", Server{})
}
