package websocket

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

func (k *Kraken) handleEvent(msg []byte) error {
	var event EventType
	if err := json.Unmarshal(msg, &event); err != nil {
		return err
	}

	switch event.Event {
	case EventPong:
		return k.handleEventPong(msg)
	case EventSystemStatus:
		return k.handleEventSystemStatus(msg)
	case EventSubscriptionStatus:
		return k.handleEventSubscriptionStatus(msg)
	case EventCancelOrderStatus:
		return k.handleEventCancelOrderStatus(msg)
	case EventAddOrderStatus:
		return k.handleEventAddOrderStatus(msg)
	case EventCancelAllStatus:
		return k.handleEventCancellAllStatus(msg)
	case EventCancelAllOrdersAfter:
		return k.handleEventCancellAllOrdersAfter(msg)
	case EventHeartbeat:
		return k.handleHeartbeat(msg)
	default:
		log.Warnf("unknown event: %s", msg)
	}
	return nil
}

func (k *Kraken) handleEventPong(data []byte) error {
	var pong PongResponse
	return json.Unmarshal(data, &pong)
}

func (k *Kraken) handleEventSystemStatus(data []byte) error {
	var systemStatus SystemStatus
	if err := json.Unmarshal(data, &systemStatus); err != nil {
		return err
	}
	k.msg <- Update{
		ChannelName: EventSystemStatus,
		Data:        systemStatus,
	}
	return nil
}

func (k *Kraken) handleHeartbeat(data []byte) error {
	k.msg <- Update{
		ChannelName: EventHeartbeat,
		Data: Heartbeat{
			Timestamp: time.Now().UTC(),
		},
	}
	return nil
}
func (k *Kraken) handleEventSubscriptionStatus(data []byte) error {
	var status SubscriptionStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return err
	}

	k.msg <- Update{
		ChannelName: EventSubscriptionStatus,
		Data:        status,
	}

	if status.Status != SubscriptionStatusError {
		if status.Status == SubscriptionStatusSubscribed {
			k.subscriptions[status.ChannelID] = &status
		} else if status.Status == SubscriptionStatusUnsubscribed {
			delete(k.subscriptions, status.ChannelID)
		}
	}
	return nil
}

func (k *Kraken) handleEventCancelOrderStatus(data []byte) error {
	var cancelOrderResponse CancelOrderResponse
	if err := json.Unmarshal(data, &cancelOrderResponse); err != nil {
		return err
	}

	switch cancelOrderResponse.Status {
	case StatusError:
		log.Errorf(cancelOrderResponse.ErrorMessage)
	case StatusOK:
		log.Debug(" Order successfully cancelled")
		k.msg <- Update{
			ChannelName: EventCancelOrder,
			Data:        cancelOrderResponse,
		}
	default:
		log.Errorf("Unknown status: %s", cancelOrderResponse.Status)
	}
	return nil
}

func (k *Kraken) handleEventAddOrderStatus(data []byte) error {
	var addOrderResponse AddOrderResponse
	if err := json.Unmarshal(data, &addOrderResponse); err != nil {
		return err
	}

	switch addOrderResponse.Status {
	case StatusError:
		log.Errorf(addOrderResponse.ErrorMessage)
	case StatusOK:
		log.Debug("Order successfully sent")
		k.msg <- Update{
			ChannelName: EventAddOrder,
			Data:        addOrderResponse,
		}
	default:
		log.Errorf("Unknown status: %s", addOrderResponse.Status)
	}
	return nil
}

func (k *Kraken) handleEventCancellAllStatus(data []byte) error {
	var cancelAllResponse CancelAllResponse
	if err := json.Unmarshal(data, &cancelAllResponse); err != nil {
		return err
	}

	switch cancelAllResponse.Status {
	case StatusError:
		log.Errorf(cancelAllResponse.ErrorMessage)
	case StatusOK:
		log.Debugf("%d orders cancelled", cancelAllResponse.Count)
		k.msg <- Update{
			ChannelName: EventCancelAllStatus,
			Data:        cancelAllResponse,
		}
	default:
		log.Errorf("Unknown status: %s", cancelAllResponse.Status)
	}
	return nil
}

func (k *Kraken) handleEventCancellAllOrdersAfter(data []byte) error {
	var cancelAllResponse CancelAllOrdersAfterResponse
	if err := json.Unmarshal(data, &cancelAllResponse); err != nil {
		return err
	}

	switch cancelAllResponse.Status {
	case StatusError:
		log.Errorf(cancelAllResponse.ErrorMessage)
	case StatusOK:
		k.msg <- Update{
			ChannelName: EventCancelAllOrdersAfter,
			Data:        cancelAllResponse,
		}
	default:
		log.Errorf("Unknown status: %s", cancelAllResponse.Status)
	}
	return nil
}
