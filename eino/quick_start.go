package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/schema"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run quick_start.go \"你的问题\" ")
		return
	}

	query := os.Args[1]
	ctx := context.Background()

	cm, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: "http://localhost:11434",
		Model:   "llama3.2",
	})

	if err != nil {
		panic(err)
	}

	messages := []*schema.Message{
		schema.SystemMessage("你是一个有帮助的助手"),
		schema.UserMessage(query),
	}

	stream, err := cm.Stream(ctx, messages)
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	fmt.Print("🤖: ")
	for {
		frame, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		if frame != nil {
			fmt.Println(frame.Content)
		}
	}
	fmt.Println()
}
