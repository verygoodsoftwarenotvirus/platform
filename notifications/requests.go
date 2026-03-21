package notifications

const (
	// MobileNotificationRequestTypeHouseholdInvitationAccepted indicates a household invitation was accepted.
	MobileNotificationRequestTypeHouseholdInvitationAccepted = "household_invitation_accepted"

	// ExcludedUserIDContextKey is the key used in MobileNotificationRequest.Context for the user ID to exclude from recipients.
	ExcludedUserIDContextKey = "excludedUserID"
)

// MobileNotificationRequest is the generic message payload for mobile push notifications.
// RequestType determines which handler processes the request; schedulers format the message.
type MobileNotificationRequest struct {
	Context          map[string]string `json:"context,omitempty"`
	BadgeCount       *int              `json:"badgeCount,omitempty"`
	RequestType      string            `json:"requestType"`
	Title            string            `json:"title"`
	Body             string            `json:"body"`
	TestID           string            `json:"testID,omitempty"`
	RecipientUserIDs []string          `json:"recipientUserIDs"`
}
