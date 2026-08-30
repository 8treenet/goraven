/*
 * Copyright 2026 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pruning_test

import (
	"context"
	"fmt"
	"goraven/backend/repository"

	"goraven/core/middleware/pruning"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func ExampleNew() {

	// 创建配置
	config := &pruning.Config{
		TokenThreshold:       160000, // Token阈值
		MaxToolResultLength:  2000,   // 最大工具结果长度
		HeadLength:           1000,   // 截断时保留头部1000字符
		TailLength:           1000,   // 截断时保留尾部1000字符
	}

	// 创建中间件
	mw, err := pruning.New(repository.NewDefaultSystemConfig(), config)
	if err != nil {
		panic(err)
	}

	fmt.Println("Pruning middleware created successfully")

	// 使用中间件
	// agent := adk.NewChatModelAgent(
	//     adk.WithModel(model),
	//     adk.WithMiddleware(mw),
	// )
	_ = mw
}

func ExampleConfig_customTokenCounter() {

	// 使用自定义token计数器
	config := &pruning.Config{
		TokenThreshold:      100000,
		MaxToolResultLength: 3000,
		HeadLength:          1500,
		TailLength:          1500,
		TokenCounter: func(ctx context.Context, msgs []adk.Message, tools []*schema.ToolInfo) (int64, error) {
			// 使用更精确的token计数方法
			var total int64
			for _, msg := range msgs {
				// 这里可以使用tiktoken等库进行精确计数
				total += int64(len(msg.Content))
			}
			return total, nil
		},
	}

	mw, err := pruning.New(repository.NewDefaultSystemConfig(), config)
	if err != nil {
		panic(err)
	}

	fmt.Println("Pruning middleware with custom token counter created successfully")
	_ = mw
}
