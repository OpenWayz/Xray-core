package connlimit

import (
	"sync"
	"sync/atomic"
	"time"
)

var (
	userConnCount sync.Map // key: email, value: *atomic.Int32
	maxConn int32 = 500    // 目前为固定上限
	once    sync.Once      // 确保后台协程只启动一次
)

// 显式启动清理任务（由 xray-core 的入口调用）
func StartCleanupTask() {
	once.Do(func() {
		go cleanupInactiveUsers()
	})
}
// incConn: 增加计数并判断是否超过限制
func IncConncustom(email string) bool {
	if email == "" {
		return true
	}
	v, _ := userConnCount.LoadOrStore(email, new(atomic.Int32))
	counter := v.(*atomic.Int32)
	n := counter.Add(1)
	if n > maxConn {
		counter.Add(-1)
		return false
	}
	return true
}
// decConn: 连接结束时递减计数
func DecConncustom(email string) {
	if email == "" {
		return
	}
	if v, ok := userConnCount.Load(email); ok {
		c := v.(*atomic.Int32)
		c.Add(-1) // 延迟清理交由后台协程
	}
}

func cleanupInactiveUsers() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		userConnCount.Range(func(key, value any) bool {
			if value.(*atomic.Int32).Load() <= 0 {
				userConnCount.Delete(key)
			}
			return true
		})
	}
}