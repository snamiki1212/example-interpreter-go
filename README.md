# interpreter go

- [O'Reilly Japan - Go 言語でつくるインタプリタ](https://www.oreilly.co.jp/books/9784873118222/) のソースコード
- グランドフィナーレまでの分

## Getting Started

```zsh
# run
$ go run main.go

# test
$ go test ./ast
$ go test ./evaluator
$ go test ./lexer
$ go test ./object
$ go test ./parser
```

## MEMO

### 全体の流れ

![本｜Goインタプリタ-2](https://user-images.githubusercontent.com/26793088/155047624-5788f757-5028-46f3-94b4-752325657789.jpg)

### `p.parseProgram`周りの流れ

![本｜Goインタプリタ-3](https://user-images.githubusercontent.com/26793088/155047621-f09ffae1-192b-43de-9b32-1f0864ed4158.jpg)

### `parseStatement`周りの流れ

![本｜Goインタプリタ-4](https://user-images.githubusercontent.com/26793088/155047615-21a870e4-b863-4710-82b9-ac56c504da90.jpg)
