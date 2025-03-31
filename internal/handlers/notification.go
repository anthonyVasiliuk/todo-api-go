package handlers

import (
	"context"
	"fmt"
	"time"
	"todo-api/pkg/logger"
)

func notifyUser(ctx context.Context, userID int, taskTitle string, ch chan<- string) {
	select {
	case <-time.After(2 * time.Second):
		logger.Log.Infof("Уведомление отправлено пользователю %d: задача '%s' создана", userID, taskTitle)
		ch <- fmt.Sprintf("Notification sent for task '%s'", taskTitle)
	case <-ctx.Done():
		logger.Log.Warnf("Уведомление для пользователя %d отменено: %v", userID, ctx.Err())
		ch <- fmt.Sprintf("Notification cancelled: %v", ctx.Err())
	}
}
