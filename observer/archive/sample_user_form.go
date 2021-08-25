package main

import (
	"fmt"
	"html/template"
	"net/http"
)

// 入力フォーム画面
func HandlerUserForm(w http.ResponseWriter, r *http.Request) {
	// テンプレートをパースする
	tpl := template.Must(template.ParseFiles("templates/user-form.gtpl"))

	// テンプレートに出力する値をマップにセット
	values := map[string]string{}

	// マップを展開してテンプレートを出力する
	if err := tpl.ExecuteTemplate(w, "user-form.gtpl", values); err != nil {
		fmt.Println(err)
	}
}

// 入力内容の確認画面
func HandlerUserConfirm(w http.ResponseWriter, req *http.Request) {
	// テンプレートをパースする
	tpl := template.Must(template.ParseFiles("templates/user-confirm.gtpl"))

	// テンプレートに出力する値をマップにセット
	values := map[string]string{
		"account": req.FormValue("account"),
		"name":    req.FormValue("name"),
		"passwd":  req.FormValue("passwd"),
	}

	// マップを展開してテンプレートを出力する
	if err := tpl.ExecuteTemplate(w, "user-confirm.gtpl", values); err != nil {
		fmt.Println(err)
	}
}

func main() {

	// "user-form"へのリクエストを関数で処理する
	http.HandleFunc("/user-form", HandlerUserForm)

	// "user-confirm"へのリクエストを関数で処理する
	http.HandleFunc("/user-confirm", HandlerUserConfirm)

	// サーバーを起動
	http.ListenAndServe(":8080", nil)
}
