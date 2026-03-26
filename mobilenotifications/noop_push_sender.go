package mobilenotifications

import "context"

// NoopPushNotificationSender is a no-op implementation of PushNotificationSender.
// It does not send any push notifications; used when APNs/FCM is not yet integrated.
type NoopPushNotificationSender struct{}

// SendPush is a no-op; it does not send any notifications.
func (n *NoopPushNotificationSender) SendPush(_ context.Context, _, _ string, _ PushMessage) error {
	return nil
}
