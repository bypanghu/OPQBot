package main

import (
	"testing"
)

/*
* @Author <bypanghu> (bypanghu@163.com)
* @Date 2023/8/3 16:46
**/

func TestHandleRiBao(t *testing.T) {
	t.Run("测试生成日报图片", func(t *testing.T) {
		HandleRiBao()
	})
}

func TestFetchTeck(t *testing.T) {
	t.Run("测试获取 科技资讯", func(t *testing.T) {
		FetchTeck()
		// if (err != nil) != tt.wantErr {
		// 	t.Errorf("FetchTeck() error = %v, wantErr %v", err, tt.wantErr)
		// 	return
		// }
		// if !reflect.DeepEqual(got, tt.want) {
		// 	t.Errorf("FetchTeck() got = %v, want %v", got, tt.want)
		// }
	})
}
