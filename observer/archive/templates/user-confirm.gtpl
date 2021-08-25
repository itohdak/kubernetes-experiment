<!DOCTYPE html>
<html>
<body>
  <h2>登録内容の確認</h2>
  <table>
    <tr>
      <td>アカウント名：</td>
      <td>{{.account}}</td>
    </tr>
    <tr>
      <td>お名前：</td>
      <td>{{.name}}</td>
    </tr>
    <tr>
      <td>パスワード：</td>
      <td>{{.passwd}}</td>
    </tr>
  </table>
  <button type="button" onclick="history.back()">入力画面に戻る</button>
</body>
</html>
