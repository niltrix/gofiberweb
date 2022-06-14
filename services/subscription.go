package services

type SubscriptionService interface {
	GetSubscription() error
	CreateSubscription() error
}
