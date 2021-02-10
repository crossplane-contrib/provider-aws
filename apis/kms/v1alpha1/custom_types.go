package v1alpha1

// CustomKeyParameters are custom parameters for Key.
type CustomKeyParameters struct {
	// Specifies whether the CMK is enabled.
	Enabled *bool `json:"enabled,omitempty"`

	// Specifies how many days the Key is retained when scheduled for deletion. Defaults to 30 days.
	PendingWindowInDays *int64 `json:"pendingWindowInDays,omitempty"`
}
