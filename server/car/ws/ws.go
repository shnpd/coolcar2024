package ws

import (
	"context"
	"coolcar/car/mq"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Handler creates a websocket http handler.
func Handler(u *websocket.Upgrader, sub mq.Subscriber, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := u.Upgrade(w, r, nil)
		if err != nil {
			logger.Warn("cannot upgrade", zap.Error(err))
			return
		}
		defer c.Close()

		msgs, cleanUp, err := sub.Subscribe(context.Background())
		defer cleanUp()
		if err != nil {
			logger.Error("cannot subscribe", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		done := make(chan struct{})
		go func() {
			for {
				_, _, err := c.ReadMessage()
				if err != nil {
					fmt.Printf("read error: %v\n", err)
					// 如果不是正常错误则打印错误信息
					if !websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
						logger.Warn("unexpected read error", zap.Error(err))
					}
					// 通知主goroutine关闭
					done <- struct{}{}
					break
				}
			}
		}()

		for {
			// 如果从rabbitmq中接收到消息则发送
			select {
			case msg := <-msgs:
				err := c.WriteJSON(msg)
				if err != nil {
					logger.Warn("cannot write message", zap.Error(err))
				}
			case <-done:
				return
			}

		}
	}

}
