package pos

import (
	"context"
	"coolcar/car/mq/amqpclt"
	coolenvpb "coolcar/shared/coolenv"
	"encoding/json"

	"go.uber.org/zap"
)

type Subscriber struct {
	Sub    *amqpclt.Subscriber
	Logger *zap.Logger
}

func (s *Subscriber) Subscribe(c context.Context) (ch chan *coolenvpb.CarPosUpdate, cleanUp func(), err error) {
	msgCh, cleanUp, err := s.Sub.SubscribeRaw(c)
	if err != nil {
		return nil, cleanUp, err
	}

	posCh := make(chan *coolenvpb.CarPosUpdate)
	go func() {
		for msg := range msgCh {
			var pos coolenvpb.CarPosUpdate
			err := json.Unmarshal(msg.Body, &pos)
			if err != nil {
				s.Logger.Error("cannot unmarshal", zap.Error(err))
				continue
			}
			posCh <- &pos
		}
		// 上面的for循环会不断从msgCh中获取消息，直到msgCh关闭退出for循环，这时候关闭carCh
		close(posCh)
	}()
	return posCh, cleanUp, nil
}
